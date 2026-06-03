package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

const (
	defaultRedisDB  = 5
	keyPrefix       = "chimoney"
	subAccountsIdx  = "chimoney:subaccounts"
	scanPageSize    = 100
)

// ChimoneySnapshot holds all exportable data from a mockchimoney Redis DB.
type ChimoneySnapshot struct {
	SubAccounts []models.SubAccount `json:"sub_accounts"`
	Payments    []models.Payment    `json:"payments"`
	Payouts     []models.Payout     `json:"payouts"`
}

func main() {
	// Global flags
	fs := flag.NewFlagSet("mockchimoney-data", flag.ExitOnError)
	valkeyURL := fs.String("valkey-url", envOr("MOCKCHIMONEY_REDIS_URL", ""), "Valkey/Redis URL (default: $MOCKCHIMONEY_REDIS_URL)")
	db := fs.Int("db", envIntOr("MOCKCHIMONEY_REDIS_DB", defaultRedisDB), "Redis DB number (default: $MOCKCHIMONEY_REDIS_DB or 5)")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: mockchimoney-data [--valkey-url URL] [--db N] <command> [options]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  export  --output FILE   Export snapshot to FILE (default: stdout)")
		fmt.Fprintln(os.Stderr, "  import  --input FILE    Import snapshot from FILE (default: stdin)")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}

	if len(os.Args) < 2 {
		fs.Usage()
		os.Exit(1)
	}

	// Parse flags up to the subcommand.
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	if *valkeyURL == "" {
		fmt.Fprintln(os.Stderr, "error: --valkey-url is required (or set MOCKCHIMONEY_REDIS_URL)")
		os.Exit(1)
	}

	client, err := newClient(*valkeyURL, *db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := context.Background()

	switch args[0] {
	case "export":
		exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
		output := exportCmd.String("output", "", "Output file (default: stdout)")
		if err := exportCmd.Parse(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		var w io.Writer = os.Stdout
		if *output != "" {
			f, err := os.Create(*output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening output file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			w = f
		}

		if err := runExport(ctx, client, w); err != nil {
			fmt.Fprintf(os.Stderr, "export error: %v\n", err)
			os.Exit(1)
		}

	case "import":
		importCmd := flag.NewFlagSet("import", flag.ExitOnError)
		input := importCmd.String("input", "", "Input file (default: stdin)")
		if err := importCmd.Parse(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		var r io.Reader = os.Stdin
		if *input != "" {
			f, err := os.Open(*input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening input file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			r = f
		}

		if err := runImport(ctx, client, r); err != nil {
			fmt.Fprintf(os.Stderr, "import error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		fs.Usage()
		os.Exit(1)
	}
}

// runExport reads all data from Redis and writes a JSON snapshot to w.
func runExport(ctx context.Context, client *redis.Client, w io.Writer) error {
	snap := ChimoneySnapshot{}
	var err error

	snap.SubAccounts, err = exportSubAccounts(ctx, client)
	if err != nil {
		return fmt.Errorf("export sub-accounts: %w", err)
	}

	snap.Payments, err = exportPayments(ctx, client)
	if err != nil {
		return fmt.Errorf("export payments: %w", err)
	}

	snap.Payouts, err = exportPayouts(ctx, client)
	if err != nil {
		return fmt.Errorf("export payouts: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	return nil
}

// exportSubAccounts uses SMEMBERS on the index key, then GET per ID.
func exportSubAccounts(ctx context.Context, client *redis.Client) ([]models.SubAccount, error) {
	ids, err := client.SMembers(ctx, subAccountsIdx).Result()
	if err != nil {
		return nil, fmt.Errorf("smembers %s: %w", subAccountsIdx, err)
	}

	accounts := make([]models.SubAccount, 0, len(ids))
	for _, id := range ids {
		key := fmt.Sprintf("%s:subaccount:%s", keyPrefix, id)
		raw, err := client.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get %s: %w", key, err)
		}
		var a models.SubAccount
		if err := json.Unmarshal([]byte(raw), &a); err != nil {
			return nil, fmt.Errorf("decode %s: %w", key, err)
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

// exportPayments scans chimoney:payment:* keeping only keys whose suffix has no colon.
func exportPayments(ctx context.Context, client *redis.Client) ([]models.Payment, error) {
	pattern := fmt.Sprintf("%s:payment:*", keyPrefix)
	prefixStrip := fmt.Sprintf("%s:payment:", keyPrefix)

	keys, err := scanKeys(ctx, client, pattern)
	if err != nil {
		return nil, err
	}

	var payments []models.Payment
	for _, key := range keys {
		suffix := strings.TrimPrefix(key, prefixStrip)
		if strings.Contains(suffix, ":") {
			continue // skip chimoney:payment:chiref:* style keys if they ever appear
		}
		raw, err := client.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get %s: %w", key, err)
		}
		var p models.Payment
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return nil, fmt.Errorf("decode %s: %w", key, err)
		}
		payments = append(payments, p)
	}
	return payments, nil
}

// exportPayouts scans chimoney:payout:* keeping only keys whose suffix has no colon
// (this correctly excludes chimoney:payout:chiref:* entries).
func exportPayouts(ctx context.Context, client *redis.Client) ([]models.Payout, error) {
	pattern := fmt.Sprintf("%s:payout:*", keyPrefix)
	prefixStrip := fmt.Sprintf("%s:payout:", keyPrefix)

	keys, err := scanKeys(ctx, client, pattern)
	if err != nil {
		return nil, err
	}

	var payouts []models.Payout
	for _, key := range keys {
		suffix := strings.TrimPrefix(key, prefixStrip)
		if strings.Contains(suffix, ":") {
			continue // excludes chimoney:payout:chiref:* lookup keys
		}
		raw, err := client.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get %s: %w", key, err)
		}
		var p models.Payout
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return nil, fmt.Errorf("decode %s: %w", key, err)
		}
		payouts = append(payouts, p)
	}
	return payouts, nil
}

// scanKeys returns all Redis keys matching the given glob pattern.
func scanKeys(ctx context.Context, client *redis.Client, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		batch, next, err := client.Scan(ctx, cursor, pattern, scanPageSize).Result()
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", pattern, err)
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// runImport flushes the DB and loads the snapshot from r.
func runImport(ctx context.Context, client *redis.Client, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	var snap ChimoneySnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}

	// 1. Flush the database.
	if err := client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flushdb: %w", err)
	}

	// 2. Import sub-accounts.
	for _, a := range snap.SubAccounts {
		key := fmt.Sprintf("%s:subaccount:%s", keyPrefix, a.ID)
		encoded, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("encode sub-account %s: %w", a.ID, err)
		}
		if err := client.Set(ctx, key, encoded, 0).Err(); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
		if err := client.SAdd(ctx, subAccountsIdx, a.ID).Err(); err != nil {
			return fmt.Errorf("sadd %s %s: %w", subAccountsIdx, a.ID, err)
		}
	}

	// 3. Import payments.
	for _, p := range snap.Payments {
		key := fmt.Sprintf("%s:payment:%s", keyPrefix, p.IssueID)
		encoded, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("encode payment %s: %w", p.IssueID, err)
		}
		if err := client.Set(ctx, key, encoded, 0).Err(); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
	}

	// 4. Import payouts and rebuild chiref lookup index.
	for _, p := range snap.Payouts {
		key := fmt.Sprintf("%s:payout:%s", keyPrefix, p.IssueID)
		encoded, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("encode payout %s: %w", p.IssueID, err)
		}
		if err := client.Set(ctx, key, encoded, 0).Err(); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
		// Rebuild the chiref → issueID lookup key.
		if p.ChiRef != "" {
			chiRefKey := fmt.Sprintf("%s:payout:chiref:%s", keyPrefix, p.ChiRef)
			if err := client.Set(ctx, chiRefKey, p.IssueID, 0).Err(); err != nil {
				return fmt.Errorf("set %s: %w", chiRefKey, err)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "imported %d sub-account(s), %d payment(s), %d payout(s)\n",
		len(snap.SubAccounts), len(snap.Payments), len(snap.Payouts))
	return nil
}

// newClient parses the URL, overrides the DB, and pings the server.
func newClient(rawURL string, db int) (*redis.Client, error) {
	opt, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid valkey URL: %w", err)
	}
	opt.DB = db

	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis/Valkey: %w", err)
	}
	return client, nil
}

// envOr returns the value of the environment variable key or fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envIntOr returns the integer value of the environment variable key or fallback.
func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	if _, err := fmt.Sscan(v, &n); err != nil {
		return fallback
	}
	return n
}
