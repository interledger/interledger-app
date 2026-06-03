package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
)

// GatehubSnapshot holds all exported mockgatehub data.
type GatehubSnapshot struct {
	Users                  []*models.User                    `json:"users"`
	Organizations          []*models.Organization            `json:"organizations"`
	Wallets                []*models.Wallet                  `json:"wallets"`
	Transactions           []*models.Transaction             `json:"transactions"`
	Balances               map[string]map[string]float64     `json:"balances"`
	Customers              []*models.Customer                `json:"customers"`
	Accounts               []*models.Account                 `json:"accounts"`
	CustomerAddresses      []*models.CustomerDeliveryAddress `json:"customer_addresses"`
	Cards                  []*models.Card                    `json:"cards"`
	CardLimits             map[string][]models.CardLimit     `json:"card_limits"`
	CardTransactions       []*models.CardTransaction         `json:"card_transactions"`
	CardTransactionsByCard map[string][]string               `json:"card_transactions_by_card"`
	ThreeDSChallenges      []*models.ThreeDSChallenge        `json:"three_ds_challenges"`
}

// Envelope wraps the snapshot with metadata.
type Envelope struct {
	Version    string          `json:"version"`
	Service    string          `json:"service"`
	ExportedAt string          `json:"exported_at"`
	Data       json.RawMessage `json:"data"`
}

// scanKeys iterates all cursor pages for the given pattern and returns every matching key.
func scanKeys(ctx context.Context, client *redis.Client, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		batch, next, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("SCAN %q: %w", pattern, err)
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// getJSON fetches a string key from Redis and JSON-unmarshals it into dest.
func getJSON(ctx context.Context, client *redis.Client, key string, dest interface{}) error {
	data, err := client.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("GET %q: %w", key, err)
	}
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("unmarshal %q: %w", key, err)
	}
	return nil
}

// exportData reads all relevant keys from the Valkey/Redis instance and builds a snapshot.
func exportData(ctx context.Context, client *redis.Client) (*GatehubSnapshot, error) {
	snap := &GatehubSnapshot{
		Balances:               make(map[string]map[string]float64),
		CardLimits:             make(map[string][]models.CardLimit),
		CardTransactionsByCard: make(map[string][]string),
	}

	// --- Users ---
	// Key: user:{uuid}  — suffix must not contain a colon (excludes user:{id}:wallets etc.)
	userKeys, err := scanKeys(ctx, client, "user:*")
	if err != nil {
		return nil, err
	}
	for _, k := range userKeys {
		// Strip "user:" prefix and check that the remainder has no colon (primary key only).
		suffix := strings.TrimPrefix(k, "user:")
		if strings.Contains(suffix, ":") {
			continue
		}
		var u models.User
		if err := getJSON(ctx, client, k, &u); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.Users = append(snap.Users, &u)
	}

	// --- Organizations ---
	// Key: organization:{id}
	orgKeys, err := scanKeys(ctx, client, "organization:*")
	if err != nil {
		return nil, err
	}
	for _, k := range orgKeys {
		var org models.Organization
		if err := getJSON(ctx, client, k, &org); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.Organizations = append(snap.Organizations, &org)
	}

	// --- Wallets ---
	// Key: wallet:{xrpl-address}  — suffix has no colon.
	walletKeys, err := scanKeys(ctx, client, "wallet:*")
	if err != nil {
		return nil, err
	}
	for _, k := range walletKeys {
		suffix := strings.TrimPrefix(k, "wallet:")
		if strings.Contains(suffix, ":") {
			continue
		}
		var w models.Wallet
		if err := getJSON(ctx, client, k, &w); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.Wallets = append(snap.Wallets, &w)
	}

	// --- Transactions ---
	// Key: tx:{uuid}  — all tx:* are primary.
	txKeys, err := scanKeys(ctx, client, "tx:*")
	if err != nil {
		return nil, err
	}
	for _, k := range txKeys {
		var tx models.Transaction
		if err := getJSON(ctx, client, k, &tx); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.Transactions = append(snap.Transactions, &tx)
	}

	// --- Balances ---
	// Key: balance:{userID}:{currency}
	balanceKeys, err := scanKeys(ctx, client, "balance:*")
	if err != nil {
		return nil, err
	}
	for _, k := range balanceKeys {
		// k = "balance:{userID}:{currency}"
		parts := strings.SplitN(k, ":", 3)
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "warning: unexpected balance key format %q, skipping\n", k)
			continue
		}
		userID := parts[1]
		currency := parts[2]
		raw, err := client.Get(ctx, k).Result()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: GET %s: %v\n", k, err)
			continue
		}
		val, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: parse float for %s: %v\n", k, err)
			continue
		}
		if snap.Balances[userID] == nil {
			snap.Balances[userID] = make(map[string]float64)
		}
		snap.Balances[userID][currency] = val
	}

	// --- Customers ---
	// Key: customer:{uuid}  — suffix must be exactly 36 chars (UUID, no colon).
	// Excludes: customer:source:*, customer:{id}:accounts, customer:{id}:addresses, customer:{id}:cards, customer:address:*
	customerKeys, err := scanKeys(ctx, client, "customer:*")
	if err != nil {
		return nil, err
	}
	for _, k := range customerKeys {
		suffix := strings.TrimPrefix(k, "customer:")
		// Primary customer keys have a suffix that is exactly a UUID (36 chars, no colon).
		if strings.Contains(suffix, ":") || len(suffix) != 36 {
			continue
		}
		var c models.Customer
		if err := getJSON(ctx, client, k, &c); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.Customers = append(snap.Customers, &c)
	}

	// --- Customer delivery addresses ---
	// Key: customer:address:{uuid}
	addrKeys, err := scanKeys(ctx, client, "customer:address:*")
	if err != nil {
		return nil, err
	}
	for _, k := range addrKeys {
		var addr models.CustomerDeliveryAddress
		if err := getJSON(ctx, client, k, &addr); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.CustomerAddresses = append(snap.CustomerAddresses, &addr)
	}

	// --- Accounts ---
	// Key: account:{uuid}  — suffix has no colon.
	accountKeys, err := scanKeys(ctx, client, "account:*")
	if err != nil {
		return nil, err
	}
	for _, k := range accountKeys {
		suffix := strings.TrimPrefix(k, "account:")
		if strings.Contains(suffix, ":") {
			continue
		}
		var a models.Account
		if err := getJSON(ctx, client, k, &a); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.Accounts = append(snap.Accounts, &a)
	}

	// --- Cards ---
	// Key: card:{uuid}  — suffix has no colon (excludes card:{id}:limits and card:{id}:transactions).
	cardKeys, err := scanKeys(ctx, client, "card:*")
	if err != nil {
		return nil, err
	}
	for _, k := range cardKeys {
		suffix := strings.TrimPrefix(k, "card:")
		if strings.Contains(suffix, ":") {
			continue
		}
		var card models.Card
		if err := getJSON(ctx, client, k, &card); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.Cards = append(snap.Cards, &card)
	}

	// --- Card limits ---
	// Key: card:{uuid}:limits
	cardLimitsKeys, err := scanKeys(ctx, client, "card:*:limits")
	if err != nil {
		return nil, err
	}
	for _, k := range cardLimitsKeys {
		// Extract cardID: strip "card:" prefix and ":limits" suffix.
		withoutPrefix := strings.TrimPrefix(k, "card:")
		cardID := strings.TrimSuffix(withoutPrefix, ":limits")
		var limits []models.CardLimit
		if err := getJSON(ctx, client, k, &limits); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.CardLimits[cardID] = limits
	}

	// --- Card transactions ---
	// Key: cardtx:{uuid}  — all cardtx:* are primary.
	cardTxKeys, err := scanKeys(ctx, client, "cardtx:*")
	if err != nil {
		return nil, err
	}
	for _, k := range cardTxKeys {
		var ctx2 models.CardTransaction
		if err := getJSON(ctx, client, k, &ctx2); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		snap.CardTransactions = append(snap.CardTransactions, &ctx2)
	}

	// --- Card-tx-by-card lists ---
	// Key: card:{uuid}:transactions
	cardTxListKeys, err := scanKeys(ctx, client, "card:*:transactions")
	if err != nil {
		return nil, err
	}
	for _, k := range cardTxListKeys {
		withoutPrefix := strings.TrimPrefix(k, "card:")
		cardID := strings.TrimSuffix(withoutPrefix, ":transactions")
		ids, err := client.LRange(ctx, k, 0, -1).Result()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: LRANGE %s: %v\n", k, err)
			continue
		}
		if len(ids) > 0 {
			snap.CardTransactionsByCard[cardID] = ids
		}
	}

	// --- 3DS challenges ---
	// Key: 3ds:challenge:{txID}  — only export non-expired challenges.
	threeDSKeys, err := scanKeys(ctx, client, "3ds:challenge:*")
	if err != nil {
		return nil, err
	}
	for _, k := range threeDSKeys {
		var ch models.ThreeDSChallenge
		if err := getJSON(ctx, client, k, &ch); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", k, err)
			continue
		}
		// Only export if the challenge has not yet timed out.
		if ch.Timeout.After(time.Now()) {
			snap.ThreeDSChallenges = append(snap.ThreeDSChallenges, &ch)
		}
	}

	return snap, nil
}

// importData flushes the database and restores all data from the snapshot.
func importData(ctx context.Context, client *redis.Client, snap *GatehubSnapshot) error {
	// 1. Flush the target database.
	if err := client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("FlushDB: %w", err)
	}

	// 2. Users.
	for _, u := range snap.Users {
		data, err := json.Marshal(u)
		if err != nil {
			return fmt.Errorf("marshal user %s: %w", u.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("user:%s", u.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET user:%s: %w", u.ID, err)
		}
		// Rebuild email→userID pointer.
		if u.Email != "" {
			if err := client.Set(ctx, fmt.Sprintf("email:%s", u.Email), u.ID, 0).Err(); err != nil {
				return fmt.Errorf("SET email:%s: %w", u.Email, err)
			}
		}
	}

	// 3. Organizations.
	for _, org := range snap.Organizations {
		data, err := json.Marshal(org)
		if err != nil {
			return fmt.Errorf("marshal organization %s: %w", org.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("organization:%s", org.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET organization:%s: %w", org.ID, err)
		}
	}

	// 4. Wallets — also rebuild user:{userID}:wallets sets.
	for _, w := range snap.Wallets {
		data, err := json.Marshal(w)
		if err != nil {
			return fmt.Errorf("marshal wallet %s: %w", w.Address, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("wallet:%s", w.Address), data, 0).Err(); err != nil {
			return fmt.Errorf("SET wallet:%s: %w", w.Address, err)
		}
		if w.UserID != "" {
			if err := client.SAdd(ctx, fmt.Sprintf("user:%s:wallets", w.UserID), w.Address).Err(); err != nil {
				return fmt.Errorf("SADD user:%s:wallets: %w", w.UserID, err)
			}
		}
	}

	// 5. Transactions.
	for _, tx := range snap.Transactions {
		data, err := json.Marshal(tx)
		if err != nil {
			return fmt.Errorf("marshal tx %s: %w", tx.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("tx:%s", tx.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET tx:%s: %w", tx.ID, err)
		}
	}

	// 6. Balances.
	for userID, currencies := range snap.Balances {
		for currency, val := range currencies {
			// Format to match strconv.FormatFloat precision used by the service.
			str := strconv.FormatFloat(val, 'f', -1, 64)
			if err := client.Set(ctx, fmt.Sprintf("balance:%s:%s", userID, currency), str, 0).Err(); err != nil {
				return fmt.Errorf("SET balance:%s:%s: %w", userID, currency, err)
			}
		}
	}

	// 7. Customers — also rebuild customer:source:{sourceID} mappings.
	for _, c := range snap.Customers {
		if c.ID == nil {
			fmt.Fprintf(os.Stderr, "warning: customer has nil ID, skipping\n")
			continue
		}
		data, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("marshal customer %s: %w", *c.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("customer:%s", *c.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET customer:%s: %w", *c.ID, err)
		}
		if c.SourceID != "" {
			if err := client.Set(ctx, fmt.Sprintf("customer:source:%s", c.SourceID), *c.ID, 0).Err(); err != nil {
				return fmt.Errorf("SET customer:source:%s: %w", c.SourceID, err)
			}
		}
	}

	// 8. Accounts — also rebuild customer:{customerID}:accounts sets.
	for _, a := range snap.Accounts {
		if a.ID == nil {
			fmt.Fprintf(os.Stderr, "warning: account has nil ID, skipping\n")
			continue
		}
		data, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("marshal account %s: %w", *a.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("account:%s", *a.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET account:%s: %w", *a.ID, err)
		}
		if a.CustomerID != nil && *a.CustomerID != "" {
			if err := client.SAdd(ctx, fmt.Sprintf("customer:%s:accounts", *a.CustomerID), *a.ID).Err(); err != nil {
				return fmt.Errorf("SADD customer:%s:accounts: %w", *a.CustomerID, err)
			}
		}
	}

	// 9. Customer delivery addresses — also rebuild customer:{customerID}:addresses sets.
	for _, addr := range snap.CustomerAddresses {
		data, err := json.Marshal(addr)
		if err != nil {
			return fmt.Errorf("marshal address %s: %w", addr.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("customer:address:%s", addr.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET customer:address:%s: %w", addr.ID, err)
		}
		if addr.CustomerID != "" {
			if err := client.SAdd(ctx, fmt.Sprintf("customer:%s:addresses", addr.CustomerID), addr.ID).Err(); err != nil {
				return fmt.Errorf("SADD customer:%s:addresses: %w", addr.CustomerID, err)
			}
		}
	}

	// 10. Cards — also rebuild customer:{customerID}:cards and account:{accountID}:cards sets.
	for _, card := range snap.Cards {
		data, err := json.Marshal(card)
		if err != nil {
			return fmt.Errorf("marshal card %s: %w", card.ID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("card:%s", card.ID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET card:%s: %w", card.ID, err)
		}
		if card.CustomerID != "" {
			if err := client.SAdd(ctx, fmt.Sprintf("customer:%s:cards", card.CustomerID), card.ID).Err(); err != nil {
				return fmt.Errorf("SADD customer:%s:cards: %w", card.CustomerID, err)
			}
		}
		if card.AccountID != "" {
			if err := client.SAdd(ctx, fmt.Sprintf("account:%s:cards", card.AccountID), card.ID).Err(); err != nil {
				return fmt.Errorf("SADD account:%s:cards: %w", card.AccountID, err)
			}
		}
	}

	// 11. Card limits.
	for cardID, limits := range snap.CardLimits {
		data, err := json.Marshal(limits)
		if err != nil {
			return fmt.Errorf("marshal card limits for %s: %w", cardID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("card:%s:limits", cardID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET card:%s:limits: %w", cardID, err)
		}
	}

	// 12. Card transactions.
	for _, ct := range snap.CardTransactions {
		data, err := json.Marshal(ct)
		if err != nil {
			return fmt.Errorf("marshal cardtx %s: %w", ct.TransactionID, err)
		}
		if err := client.Set(ctx, fmt.Sprintf("cardtx:%s", ct.TransactionID), data, 0).Err(); err != nil {
			return fmt.Errorf("SET cardtx:%s: %w", ct.TransactionID, err)
		}
	}

	// 13. Card-tx-by-card lists.
	for cardID, txIDs := range snap.CardTransactionsByCard {
		if len(txIDs) == 0 {
			continue
		}
		// Convert []string to []interface{} for RPUSH variadic.
		members := make([]interface{}, len(txIDs))
		for i, id := range txIDs {
			members[i] = id
		}
		if err := client.RPush(ctx, fmt.Sprintf("card:%s:transactions", cardID), members...).Err(); err != nil {
			return fmt.Errorf("RPUSH card:%s:transactions: %w", cardID, err)
		}
	}

	// 14. 3DS challenges — restore with TTL and rebuild user sorted sets.
	for _, ch := range snap.ThreeDSChallenges {
		data, err := json.Marshal(ch)
		if err != nil {
			return fmt.Errorf("marshal 3DS challenge %s: %w", ch.TransactionID, err)
		}
		ttl := time.Until(ch.Timeout) + 10*time.Minute
		if ttl <= 0 {
			// Should not happen (we only export non-expired), but be safe.
			ttl = 10 * time.Minute
		}
		if err := client.Set(ctx, fmt.Sprintf("3ds:challenge:%s", ch.TransactionID), data, ttl).Err(); err != nil {
			return fmt.Errorf("SET 3ds:challenge:%s: %w", ch.TransactionID, err)
		}
		// Rebuild user:{userID}:3ds:challenges sorted set (score = timeout unix timestamp).
		if ch.UserID != "" {
			score := float64(ch.Timeout.Unix())
			if err := client.ZAdd(ctx, fmt.Sprintf("user:%s:3ds:challenges", ch.UserID), redis.Z{
				Score:  score,
				Member: ch.TransactionID,
			}).Err(); err != nil {
				return fmt.Errorf("ZADD user:%s:3ds:challenges: %w", ch.UserID, err)
			}
		}
	}

	return nil
}

func connectRedis(valkeyURL string, db int) (*redis.Client, error) {
	opt, err := redis.ParseURL(valkeyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Valkey/Redis URL: %w", err)
	}
	opt.DB = db
	client := redis.NewClient(opt)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey/Redis: %w", err)
	}
	return client, nil
}

func defaultEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func defaultEnvStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <export|import> [flags]\n", os.Args[0])
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "export":
		fs := flag.NewFlagSet("export", flag.ExitOnError)
		valkeyURL := fs.String("valkey-url", defaultEnvStr("MOCKGATEHUB_REDIS_URL", "redis://localhost:6379"), "Valkey/Redis URL")
		db := fs.Int("db", defaultEnvInt("MOCKGATEHUB_REDIS_DB", 2), "Redis database number")
		output := fs.String("output", "", "Output file path (default: stdout)")
		if err := fs.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
			os.Exit(1)
		}

		client, err := connectRedis(*valkeyURL, *db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		ctx := context.Background()
		snap, err := exportData(ctx, client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "export: %v\n", err)
			os.Exit(1)
		}

		snapJSON, err := json.Marshal(snap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal snapshot: %v\n", err)
			os.Exit(1)
		}

		env := Envelope{
			Version:    "1",
			Service:    "mockgatehub",
			ExportedAt: time.Now().UTC().Format(time.RFC3339),
			Data:       json.RawMessage(snapJSON),
		}

		envJSON, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal envelope: %v\n", err)
			os.Exit(1)
		}

		var w io.Writer = os.Stdout
		if *output != "" {
			f, err := os.Create(*output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "open output file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			w = f
		}

		if _, err := w.Write(envJSON); err != nil {
			fmt.Fprintf(os.Stderr, "write output: %v\n", err)
			os.Exit(1)
		}
		// Ensure trailing newline for stdout.
		fmt.Fprintln(w)

	case "import":
		fs := flag.NewFlagSet("import", flag.ExitOnError)
		valkeyURL := fs.String("valkey-url", defaultEnvStr("MOCKGATEHUB_REDIS_URL", "redis://localhost:6379"), "Valkey/Redis URL")
		db := fs.Int("db", defaultEnvInt("MOCKGATEHUB_REDIS_DB", 2), "Redis database number")
		input := fs.String("input", "", "Input file path (default: stdin)")
		if err := fs.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
			os.Exit(1)
		}

		var r io.Reader = os.Stdin
		if *input != "" {
			f, err := os.Open(*input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "open input file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			r = f
		}

		raw, err := io.ReadAll(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read input: %v\n", err)
			os.Exit(1)
		}

		var env Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			fmt.Fprintf(os.Stderr, "unmarshal envelope: %v\n", err)
			os.Exit(1)
		}

		var snap GatehubSnapshot
		if err := json.Unmarshal(env.Data, &snap); err != nil {
			fmt.Fprintf(os.Stderr, "unmarshal snapshot: %v\n", err)
			os.Exit(1)
		}

		client, err := connectRedis(*valkeyURL, *db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		ctx := context.Background()
		if err := importData(ctx, client, &snap); err != nil {
			fmt.Fprintf(os.Stderr, "import: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "import complete (exported_at: %s)\n", env.ExportedAt)

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q — use 'export' or 'import'\n", subcommand)
		os.Exit(1)
	}
}
