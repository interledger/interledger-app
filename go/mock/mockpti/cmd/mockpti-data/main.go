package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

// WalletSnap captures a models.Wallet including the UserID that is json:"-" on the original type.
type WalletSnap struct {
	UserID         string  `json:"user_id"`
	WalletID       string  `json:"wallet_id"`
	Currency       string  `json:"currency,omitempty"`
	Reference      string  `json:"reference,omitempty"`
	CreateDateTime string  `json:"createDateTime,omitempty"`
	Balance        float64 `json:"balance"`
}

// PaymentInfoSnap captures a models.PaymentInformation including the UserID that is json:"-" on the original type.
type PaymentInfoSnap struct {
	UserID                string `json:"user_id"`
	ID                    string `json:"id,omitempty"`
	Type                  string `json:"type"`
	BankAccountNumber     string `json:"bankAccountNumber,omitempty"`
	BankAccountType       string `json:"bankAccountType,omitempty"`
	BankSwiftCode         string `json:"bankSwiftCode,omitempty"`
	BankRoutingNumber     string `json:"bankRoutingNumber,omitempty"`
	BankRoutingCheckDigit string `json:"bankRoutingCheckDigit,omitempty"`
	AccountBankName       string `json:"accountBankName,omitempty"`
}

// TxUpdateSnap captures a models.TransactionUpdate including the RequestID that is json:"-" on the original type.
type TxUpdateSnap struct {
	RequestID     string    `json:"request_id"`
	ID            string    `json:"id"`
	TransactionID string    `json:"transactionId"`
	Feedback      string    `json:"feedback"`
	Date          time.Time `json:"date"`
	ProviderName  string    `json:"providerName"`
	Payload       string    `json:"payload"`
}

// PTISnapshot is the top-level export/import structure.
type PTISnapshot struct {
	Users              []*models.User                  `json:"users"`
	Assessments        map[string][]*models.Assessment `json:"assessments"`
	Wallets            []*WalletSnap                   `json:"wallets"`
	PaymentInformation []*PaymentInfoSnap              `json:"payment_information"`
	Transactions       []*models.Transaction           `json:"transactions"`
	TransactionUpdates map[string][]*TxUpdateSnap      `json:"transaction_updates"`
}

func main() {
	// Global flags
	valkeyURL := flag.String("valkey-url", envOrDefault("MOCKPTI_REDIS_URL", "redis://localhost:6379"), "Valkey/Redis URL")
	dbStr := flag.String("db", envOrDefault("MOCKPTI_REDIS_DB", "0"), "Redis DB index")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: mockpti-data [--valkey-url URL] [--db N] <command> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  export  Export all PTI data to JSON\n")
		fmt.Fprintf(os.Stderr, "  import  Import PTI data from JSON\n\n")
		fmt.Fprintf(os.Stderr, "Global flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	db, err := strconv.Atoi(*dbStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --db value %q: %v\n", *dbStr, err)
		os.Exit(1)
	}

	client, err := newClient(*valkeyURL, db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	switch cmd {
	case "export":
		if err := runExport(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "export failed: %v\n", err)
			os.Exit(1)
		}
	case "import":
		if err := runImport(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "import failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// newClient creates a Redis client from rawURL (redis:// / rediss:// URL or plain host:port).
func newClient(rawURL string, db int) (*redis.Client, error) {
	var opt *redis.Options
	if parsed, err := redis.ParseURL(rawURL); err == nil {
		opt = parsed
	} else {
		opt = &redis.Options{Addr: rawURL, PoolSize: 5}
	}
	opt.DB = db

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}
	return client, nil
}

// scanKeys returns all keys matching the given pattern using SCAN.
func scanKeys(ctx context.Context, client *redis.Client, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("SCAN %s: %w", pattern, err)
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// exportData reads all PTI data from Valkey and returns a snapshot.
func exportData(ctx context.Context, client *redis.Client) (*PTISnapshot, error) {
	snap := &PTISnapshot{
		Users:              make([]*models.User, 0),
		Assessments:        make(map[string][]*models.Assessment),
		Wallets:            make([]*WalletSnap, 0),
		PaymentInformation: make([]*PaymentInfoSnap, 0),
		Transactions:       make([]*models.Transaction, 0),
		TransactionUpdates: make(map[string][]*TxUpdateSnap),
	}

	// --- Users ---
	userKeys, err := scanKeys(ctx, client, "pti:user:*")
	if err != nil {
		return nil, err
	}
	for _, k := range userKeys {
		data, err := client.Get(ctx, k).Bytes()
		if err != nil {
			return nil, fmt.Errorf("GET %s: %w", k, err)
		}
		var u models.User
		if err := json.Unmarshal(data, &u); err != nil {
			return nil, fmt.Errorf("unmarshal user %s: %w", k, err)
		}
		snap.Users = append(snap.Users, &u)
	}

	// --- Assessments ---
	assessmentKeys, err := scanKeys(ctx, client, "pti:assessments:*")
	if err != nil {
		return nil, err
	}
	for _, k := range assessmentKeys {
		// key = pti:assessments:{userID}
		userID := strings.TrimPrefix(k, "pti:assessments:")
		items, err := client.LRange(ctx, k, 0, -1).Result()
		if err != nil {
			return nil, fmt.Errorf("LRANGE %s: %w", k, err)
		}
		list := make([]*models.Assessment, 0, len(items))
		for _, item := range items {
			var a models.Assessment
			if err := json.Unmarshal([]byte(item), &a); err != nil {
				return nil, fmt.Errorf("unmarshal assessment in %s: %w", k, err)
			}
			list = append(list, &a)
		}
		snap.Assessments[userID] = list
	}

	// --- Wallets ---
	walletIndexKeys, err := scanKeys(ctx, client, "pti:wallets:*")
	if err != nil {
		return nil, err
	}
	for _, indexKey := range walletIndexKeys {
		// key = pti:wallets:{userID}
		userID := strings.TrimPrefix(indexKey, "pti:wallets:")
		walletIDs, err := client.SMembers(ctx, indexKey).Result()
		if err != nil {
			return nil, fmt.Errorf("SMEMBERS %s: %w", indexKey, err)
		}
		for _, walletID := range walletIDs {
			wKey := fmt.Sprintf("pti:wallet:%s:%s", userID, walletID)
			data, err := client.Get(ctx, wKey).Bytes()
			if err != nil {
				return nil, fmt.Errorf("GET %s: %w", wKey, err)
			}
			var w models.Wallet
			if err := json.Unmarshal(data, &w); err != nil {
				return nil, fmt.Errorf("unmarshal wallet %s: %w", wKey, err)
			}
			snap.Wallets = append(snap.Wallets, &WalletSnap{
				UserID:         userID,
				WalletID:       w.WalletID,
				Currency:       w.Currency,
				Reference:      w.Reference,
				CreateDateTime: w.CreateDateTime,
				Balance:        w.Balance,
			})
		}
	}

	// --- Payment Information ---
	piKeys, err := scanKeys(ctx, client, "pti:paymentinfo:*")
	if err != nil {
		return nil, err
	}
	for _, k := range piKeys {
		// key = pti:paymentinfo:{userID}:{piID}
		suffix := strings.TrimPrefix(k, "pti:paymentinfo:")
		// userID may itself contain ":" only if UUIDs — PTI uses UUIDs so safe to split on first ":"
		sepIdx := strings.Index(suffix, ":")
		if sepIdx < 0 {
			return nil, fmt.Errorf("unexpected paymentinfo key format: %s", k)
		}
		userID := suffix[:sepIdx]
		data, err := client.Get(ctx, k).Bytes()
		if err != nil {
			return nil, fmt.Errorf("GET %s: %w", k, err)
		}
		var pi models.PaymentInformation
		if err := json.Unmarshal(data, &pi); err != nil {
			return nil, fmt.Errorf("unmarshal payment info %s: %w", k, err)
		}
		snap.PaymentInformation = append(snap.PaymentInformation, &PaymentInfoSnap{
			UserID:                userID,
			ID:                    pi.ID,
			Type:                  pi.Type,
			BankAccountNumber:     pi.BankAccountNumber,
			BankAccountType:       pi.BankAccountType,
			BankSwiftCode:         pi.BankSwiftCode,
			BankRoutingNumber:     pi.BankRoutingNumber,
			BankRoutingCheckDigit: pi.BankRoutingCheckDigit,
			AccountBankName:       pi.AccountBankName,
		})
	}

	// --- Transactions ---
	txKeys, err := scanKeys(ctx, client, "pti:transaction:*")
	if err != nil {
		return nil, err
	}
	for _, k := range txKeys {
		data, err := client.Get(ctx, k).Bytes()
		if err != nil {
			return nil, fmt.Errorf("GET %s: %w", k, err)
		}
		var tx models.Transaction
		if err := json.Unmarshal(data, &tx); err != nil {
			return nil, fmt.Errorf("unmarshal transaction %s: %w", k, err)
		}
		snap.Transactions = append(snap.Transactions, &tx)
	}

	// --- Transaction Updates ---
	txUpdateKeys, err := scanKeys(ctx, client, "pti:txupdates:*")
	if err != nil {
		return nil, err
	}
	for _, k := range txUpdateKeys {
		// key = pti:txupdates:{requestID}
		requestID := strings.TrimPrefix(k, "pti:txupdates:")
		items, err := client.LRange(ctx, k, 0, -1).Result()
		if err != nil {
			return nil, fmt.Errorf("LRANGE %s: %w", k, err)
		}
		list := make([]*TxUpdateSnap, 0, len(items))
		for _, item := range items {
			var upd models.TransactionUpdate
			if err := json.Unmarshal([]byte(item), &upd); err != nil {
				return nil, fmt.Errorf("unmarshal tx update in %s: %w", k, err)
			}
			list = append(list, &TxUpdateSnap{
				RequestID:     requestID,
				ID:            upd.ID,
				TransactionID: upd.TransactionID,
				Feedback:      upd.Feedback,
				Date:          upd.Date,
				ProviderName:  upd.ProviderName,
				Payload:       upd.Payload,
			})
		}
		snap.TransactionUpdates[requestID] = list
	}

	return snap, nil
}

// runExport reads all PTI data from Valkey and writes JSON to the output file (or stdout).
func runExport(client *redis.Client, args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	output := fs.String("output", "", "Output file (default: stdout)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx := context.Background()
	snap, err := exportData(ctx, client)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	out = append(out, '\n')

	if *output == "" {
		_, err = os.Stdout.Write(out)
		return err
	}
	return os.WriteFile(*output, out, 0o644)
}

// importData restores PTI state from a snapshot.
func importData(ctx context.Context, client *redis.Client, snap *PTISnapshot, noFlush bool) error {
	// 1. Flush DB (skip if --no-flush or FlushDB is disabled on the server)
	if !noFlush {
		if err := client.FlushDB(ctx).Err(); err != nil {
			return fmt.Errorf("FlushDB: %w", err)
		}
	}

	// 2. Users
	for _, u := range snap.Users {
		data, err := json.Marshal(u)
		if err != nil {
			return fmt.Errorf("marshal user %s: %w", u.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("pti:user:%s", u.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET user %s: %w", u.ID, err)
		}
	}

	// 3. Assessments
	for userID, assessments := range snap.Assessments {
		key := fmt.Sprintf("pti:assessments:%s", userID)
		for _, a := range assessments {
			data, err := json.Marshal(a)
			if err != nil {
				return fmt.Errorf("marshal assessment for user %s: %w", userID, err)
			}
			if err := client.RPush(ctx, key, data).Err(); err != nil {
				return fmt.Errorf("RPUSH %s: %w", key, err)
			}
		}
	}

	// 4. Wallets
	for _, ws := range snap.Wallets {
		w := &models.Wallet{
			WalletID:       ws.WalletID,
			Currency:       ws.Currency,
			Reference:      ws.Reference,
			CreateDateTime: ws.CreateDateTime,
			Balance:        ws.Balance,
			// UserID is json:"-" so we set it explicitly before marshalling
			UserID: ws.UserID,
		}
		// Marshal without UserID (json:"-") — same as how the service stores it
		data, err := json.Marshal(w)
		if err != nil {
			return fmt.Errorf("marshal wallet %s/%s: %w", ws.UserID, ws.WalletID, err)
		}
		wKey := fmt.Sprintf("pti:wallet:%s:%s", ws.UserID, ws.WalletID)
		indexKey := fmt.Sprintf("pti:wallets:%s", ws.UserID)
		pipe := client.Pipeline()
		pipe.Set(ctx, wKey, data, 0)
		pipe.SAdd(ctx, indexKey, ws.WalletID)
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("SET/SADD wallet %s/%s: %w", ws.UserID, ws.WalletID, err)
		}
	}

	// 5. Payment Information
	for _, ps := range snap.PaymentInformation {
		pi := &models.PaymentInformation{
			ID:                    ps.ID,
			Type:                  ps.Type,
			BankAccountNumber:     ps.BankAccountNumber,
			BankAccountType:       ps.BankAccountType,
			BankSwiftCode:         ps.BankSwiftCode,
			BankRoutingNumber:     ps.BankRoutingNumber,
			BankRoutingCheckDigit: ps.BankRoutingCheckDigit,
			AccountBankName:       ps.AccountBankName,
			// UserID is json:"-" — set explicitly before marshalling
			UserID: ps.UserID,
		}
		// Marshal without UserID (json:"-") — same as how the service stores it
		data, err := json.Marshal(pi)
		if err != nil {
			return fmt.Errorf("marshal payment info %s/%s: %w", ps.UserID, ps.ID, err)
		}
		piKey := fmt.Sprintf("pti:paymentinfo:%s:%s", ps.UserID, ps.ID)
		if err := client.Set(ctx, piKey, data, 0).Err(); err != nil {
			return fmt.Errorf("SET %s: %w", piKey, err)
		}
	}

	// 6. Transactions
	for _, tx := range snap.Transactions {
		data, err := json.Marshal(tx)
		if err != nil {
			return fmt.Errorf("marshal transaction %s: %w", tx.RequestID, err)
		}
		txKey := fmt.Sprintf("pti:transaction:%s", tx.RequestID)
		if err := client.Set(ctx, txKey, data, 0).Err(); err != nil {
			return fmt.Errorf("SET %s: %w", txKey, err)
		}
	}

	// 7. Transaction Updates — store with RequestID omitted (json:"-" on original type)
	for requestID, updates := range snap.TransactionUpdates {
		key := fmt.Sprintf("pti:txupdates:%s", requestID)
		for _, us := range updates {
			// Reconstruct the original TransactionUpdate (RequestID will be json:"-" so omitted in marshal)
			upd := &models.TransactionUpdate{
				ID:            us.ID,
				TransactionID: us.TransactionID,
				Feedback:      us.Feedback,
				Date:          us.Date,
				ProviderName:  us.ProviderName,
				Payload:       us.Payload,
				// RequestID intentionally not set — it's json:"-" and recovered from the key
			}
			data, err := json.Marshal(upd)
			if err != nil {
				return fmt.Errorf("marshal tx update for %s: %w", requestID, err)
			}
			if err := client.RPush(ctx, key, data).Err(); err != nil {
				return fmt.Errorf("RPUSH %s: %w", key, err)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "import complete: %d users, %d assessment keys, %d wallets, %d payment infos, %d transactions, %d tx update keys\n",
		len(snap.Users),
		len(snap.Assessments),
		len(snap.Wallets),
		len(snap.PaymentInformation),
		len(snap.Transactions),
		len(snap.TransactionUpdates),
	)
	return nil
}

// runImport reads a JSON snapshot and restores it into Valkey after flushing the DB.
func runImport(client *redis.Client, args []string) error {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	input := fs.String("input", "", "Input file (default: stdin)")
	noFlush := fs.Bool("no-flush", false, "Skip FlushDB before import (useful when FlushDB is disabled)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Read snapshot
	var raw []byte
	var err error
	if *input == "" {
		raw, err = os.ReadFile("/dev/stdin")
	} else {
		raw, err = os.ReadFile(*input)
	}
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	var snap PTISnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return fmt.Errorf("unmarshal snapshot: %w", err)
	}

	ctx := context.Background()
	return importData(ctx, client, &snap, *noFlush)
}
