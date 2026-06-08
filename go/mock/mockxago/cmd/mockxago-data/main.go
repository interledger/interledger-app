package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

// XagoBalance holds available and reserved amounts for a single wallet+currency pair.
type XagoBalance struct {
	Available float64 `json:"available"`
	Reserved  float64 `json:"reserved"`
}

// XagoSnapshot is the portable representation of all persistent mockxago state.
type XagoSnapshot struct {
	SubAccounts   []*models.SubAccount              `json:"sub_accounts"`
	Beneficiaries []*models.Beneficiary             `json:"beneficiaries"`
	Transactions  []*models.Transaction             `json:"transactions"`
	Deposits      []*models.Deposit                 `json:"deposits"`
	Balances      map[string]map[string]XagoBalance `json:"balances"` // walletID → currency → balance
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  mockxago-data [--valkey-url URL] [--db N] export [--output FILE]
  mockxago-data [--valkey-url URL] [--db N] import [--input FILE]

Flags:
  --valkey-url  Redis/Valkey URL (default: $MOCKXAGO_REDIS_URL)
  --db          Database index (default: $MOCKXAGO_REDIS_DB or 0)

Subcommands:
  export  Dump persistent mockxago state to JSON (stdout or --output FILE)
  import  Restore mockxago state from JSON (stdin or --input FILE); flushes DB first
`)
	os.Exit(1)
}

func main() {
	args := os.Args[1:]

	valkeyURL := os.Getenv("MOCKXAGO_REDIS_URL")
	dbIndex := 0
	if v := os.Getenv("MOCKXAGO_REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			dbIndex = n
		}
	}

	// Parse global flags before the subcommand
	for len(args) > 0 && strings.HasPrefix(args[0], "--") {
		switch args[0] {
		case "--valkey-url":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "missing value for --valkey-url")
				usage()
			}
			valkeyURL = args[1]
			args = args[2:]
		case "--db":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "missing value for --db")
				usage()
			}
			n, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid --db value: %v\n", err)
				usage()
			}
			dbIndex = n
			args = args[2:]
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[0])
			usage()
		}
	}

	if len(args) == 0 {
		usage()
	}

	subcommand := args[0]
	args = args[1:]

	if valkeyURL == "" {
		fmt.Fprintln(os.Stderr, "redis/valkey URL is required (--valkey-url or $MOCKXAGO_REDIS_URL)")
		os.Exit(1)
	}

	client, err := newRedisClient(valkeyURL, dbIndex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := context.Background()

	switch subcommand {
	case "export":
		outputFile := ""
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--output":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "missing value for --output")
					usage()
				}
				outputFile = args[i+1]
				i++
			default:
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[i])
				usage()
			}
		}
		if err := runExport(ctx, client, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "export failed: %v\n", err)
			os.Exit(1)
		}

	case "import":
		inputFile := ""
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--input":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "missing value for --input")
					usage()
				}
				inputFile = args[i+1]
				i++
			default:
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[i])
				usage()
			}
		}
		if err := runImport(ctx, client, inputFile); err != nil {
			fmt.Fprintf(os.Stderr, "import failed: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subcommand)
		usage()
	}
}

// newRedisClient creates a Redis client from rawURL (redis:// / rediss:// URL or plain host:port).
func newRedisClient(rawURL string, db int) (*redis.Client, error) {
	var opt *redis.Options
	if parsed, err := redis.ParseURL(rawURL); err == nil {
		opt = parsed
	} else {
		opt = &redis.Options{Addr: rawURL}
	}
	opt.DB = db

	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}
	return client, nil
}

// exportData scans Redis for all persistent mockxago keys and returns a snapshot.
func exportData(ctx context.Context, client *redis.Client) (*XagoSnapshot, error) {
	snap := &XagoSnapshot{
		SubAccounts:   []*models.SubAccount{},
		Beneficiaries: []*models.Beneficiary{},
		Transactions:  []*models.Transaction{},
		Deposits:      []*models.Deposit{},
		Balances:      map[string]map[string]XagoBalance{},
	}

	// Sub-accounts: keys matching "subaccount:*" where the suffix contains no ":"
	subAccountKeys, err := scanKeysWithFilter(ctx, client, "subaccount:*", func(key string) bool {
		suffix := strings.TrimPrefix(key, "subaccount:")
		return !strings.Contains(suffix, ":")
	})
	if err != nil {
		return nil, fmt.Errorf("scanning sub-accounts: %w", err)
	}
	for _, key := range subAccountKeys {
		data, err := client.Get(ctx, key).Bytes()
		if err != nil {
			return nil, fmt.Errorf("getting sub-account %s: %w", key, err)
		}
		var sa models.SubAccount
		if err := json.Unmarshal(data, &sa); err != nil {
			return nil, fmt.Errorf("unmarshalling sub-account %s: %w", key, err)
		}
		snap.SubAccounts = append(snap.SubAccounts, &sa)
	}

	// Beneficiaries: keys matching "beneficiary:*" where the suffix contains no ":"
	beneficiaryKeys, err := scanKeysWithFilter(ctx, client, "beneficiary:*", func(key string) bool {
		suffix := strings.TrimPrefix(key, "beneficiary:")
		return !strings.Contains(suffix, ":")
	})
	if err != nil {
		return nil, fmt.Errorf("scanning beneficiaries: %w", err)
	}
	for _, key := range beneficiaryKeys {
		data, err := client.Get(ctx, key).Bytes()
		if err != nil {
			return nil, fmt.Errorf("getting beneficiary %s: %w", key, err)
		}
		var b models.Beneficiary
		if err := json.Unmarshal(data, &b); err != nil {
			return nil, fmt.Errorf("unmarshalling beneficiary %s: %w", key, err)
		}
		snap.Beneficiaries = append(snap.Beneficiaries, &b)
	}

	// Transactions: keys matching "transaction:*" where the suffix contains no ":"
	transactionKeys, err := scanKeysWithFilter(ctx, client, "transaction:*", func(key string) bool {
		suffix := strings.TrimPrefix(key, "transaction:")
		return !strings.Contains(suffix, ":")
	})
	if err != nil {
		return nil, fmt.Errorf("scanning transactions: %w", err)
	}
	for _, key := range transactionKeys {
		data, err := client.Get(ctx, key).Bytes()
		if err != nil {
			return nil, fmt.Errorf("getting transaction %s: %w", key, err)
		}
		var tx models.Transaction
		if err := json.Unmarshal(data, &tx); err != nil {
			return nil, fmt.Errorf("unmarshalling transaction %s: %w", key, err)
		}
		snap.Transactions = append(snap.Transactions, &tx)
	}

	// Deposits: use the ordered list "deposits:all"
	depositIDs, err := client.LRange(ctx, "deposits:all", 0, -1).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("reading deposits:all: %w", err)
	}
	for _, id := range depositIDs {
		data, err := client.Get(ctx, fmt.Sprintf("deposit:%s", id)).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue // already deleted
			}
			return nil, fmt.Errorf("getting deposit %s: %w", id, err)
		}
		var dep models.Deposit
		if err := json.Unmarshal(data, &dep); err != nil {
			return nil, fmt.Errorf("unmarshalling deposit %s: %w", id, err)
		}
		snap.Deposits = append(snap.Deposits, &dep)
	}

	// Balances: keys matching "balance:*" with exactly 4 colon-separated parts
	// Format: balance:{walletID}:{currency}:available|reserved
	balanceKeys, err := scanKeysWithFilter(ctx, client, "balance:*", func(key string) bool {
		parts := strings.Split(key, ":")
		if len(parts) != 4 {
			return false
		}
		last := parts[3]
		return last == "available" || last == "reserved"
	})
	if err != nil {
		return nil, fmt.Errorf("scanning balances: %w", err)
	}
	for _, key := range balanceKeys {
		parts := strings.Split(key, ":")
		// parts: ["balance", walletID, currency, "available"|"reserved"]
		walletID := parts[1]
		currency := parts[2]
		kind := parts[3]

		valStr, err := client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, fmt.Errorf("getting balance %s: %w", key, err)
		}
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing balance value %q for key %s: %w", valStr, key, err)
		}

		if snap.Balances[walletID] == nil {
			snap.Balances[walletID] = map[string]XagoBalance{}
		}
		b := snap.Balances[walletID][currency]
		switch kind {
		case "available":
			b.Available = val
		case "reserved":
			b.Reserved = val
		}
		snap.Balances[walletID][currency] = b
	}

	return snap, nil
}

// runExport exports data to a file or stdout.
func runExport(ctx context.Context, client *redis.Client, outputFile string) error {
	snap, err := exportData(ctx, client)
	if err != nil {
		return err
	}

	var out io.Writer = os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("encoding snapshot: %w", err)
	}
	return nil
}

// importData flushes the DB and restores all data from a snapshot.
func importData(ctx context.Context, client *redis.Client, snap *XagoSnapshot) error {
	if err := client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flushing database: %w", err)
	}

	pipe := client.Pipeline()

	// Sub-accounts
	for _, sa := range snap.SubAccounts {
		data, err := json.Marshal(sa)
		if err != nil {
			return fmt.Errorf("marshalling sub-account %s: %w", sa.AccountID, err)
		}
		pipe.Set(ctx, fmt.Sprintf("subaccount:%s", sa.AccountID), data, 0)
		pipe.Set(ctx, fmt.Sprintf("subaccount:wallet:%s", sa.WalletID), sa.AccountID, 0)
		if sa.DepositReferenceZAR != "" {
			pipe.Set(ctx, fmt.Sprintf("subaccount:depositref:%s", sa.DepositReferenceZAR), sa.AccountID, 0)
		}
		if sa.DepositReferenceUSD != "" {
			pipe.Set(ctx, fmt.Sprintf("subaccount:depositref:%s", sa.DepositReferenceUSD), sa.AccountID, 0)
		}
	}

	// Beneficiaries
	// Note: SaveBeneficiary uses beneficiary.AccountID for both the wallet list key and account list key.
	for _, b := range snap.Beneficiaries {
		data, err := json.Marshal(b)
		if err != nil {
			return fmt.Errorf("marshalling beneficiary %s: %w", b.ID, err)
		}
		pipe.Set(ctx, fmt.Sprintf("beneficiary:%s", b.ID), data, 0)
		pipe.RPush(ctx, fmt.Sprintf("beneficiaries:wallet:%s", b.AccountID), b.ID)
		pipe.RPush(ctx, fmt.Sprintf("beneficiaries:account:%s", b.AccountID), b.ID)
	}

	// Transactions
	for _, tx := range snap.Transactions {
		data, err := json.Marshal(tx)
		if err != nil {
			return fmt.Errorf("marshalling transaction %s: %w", tx.ID, err)
		}
		pipe.Set(ctx, fmt.Sprintf("transaction:%s", tx.ID), data, 0)
		pipe.RPush(ctx, fmt.Sprintf("transactions:account:%s", tx.AccountID), tx.ID)
	}

	// Deposits
	for _, dep := range snap.Deposits {
		data, err := json.Marshal(dep)
		if err != nil {
			return fmt.Errorf("marshalling deposit %s: %w", dep.ID, err)
		}
		pipe.Set(ctx, fmt.Sprintf("deposit:%s", dep.ID), data, 0)
		if dep.DepositReference != "" {
			pipe.Set(ctx, fmt.Sprintf("deposit:ref:%s", dep.DepositReference), dep.ID, 0)
		}
		pipe.RPush(ctx, "deposits:all", dep.ID)
	}

	// Balances
	for walletID, currencies := range snap.Balances {
		for currency, bal := range currencies {
			pipe.Set(ctx, fmt.Sprintf("balance:%s:%s:available", walletID, currency), fmt.Sprintf("%.2f", bal.Available), 0)
			pipe.Set(ctx, fmt.Sprintf("balance:%s:%s:reserved", walletID, currency), fmt.Sprintf("%.2f", bal.Reserved), 0)
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("executing import pipeline: %w", err)
	}

	return nil
}

// runImport reads a JSON snapshot from a file (or stdin) and restores it.
func runImport(ctx context.Context, client *redis.Client, inputFile string) error {
	var in io.Reader = os.Stdin
	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("opening input file: %w", err)
		}
		defer f.Close()
		in = f
	}

	var snap XagoSnapshot
	if err := json.NewDecoder(in).Decode(&snap); err != nil {
		return fmt.Errorf("decoding snapshot: %w", err)
	}

	return importData(ctx, client, &snap)
}

// scanKeysWithFilter uses SCAN to iterate over keys matching pattern and returns
// those that pass the filter predicate.
func scanKeysWithFilter(ctx context.Context, client *redis.Client, pattern string, filter func(string) bool) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		var batch []string
		var err error
		batch, cursor, err = client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range batch {
			if filter(key) {
				keys = append(keys, key)
			}
		}
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}
