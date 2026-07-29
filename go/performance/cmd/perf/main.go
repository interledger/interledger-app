// Command perf measures transaction throughput and latency against a running
// Interledger wallet backend.
//
//	perf run      -config scenarios/many-to-many.yaml [-config overlay.yaml]
//	perf validate -config scenarios/many-to-many.yaml
//	perf provision -countries za,de,us -per-country 100 -target 5000
//
// See go/performance/README.md for the scenario format and for how to point this
// at a deployed environment through a port-forward.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/interledger/interledger-app/go/performance/config"
	"github.com/interledger/interledger-app/go/performance/metrics"
	"github.com/interledger/interledger-app/go/performance/provision"
	"github.com/interledger/interledger-app/go/performance/runner"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nperf: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return errors.New("a subcommand is required")
	}

	// Interrupt cancels the run rather than killing it, so in-flight payments are
	// accounted for and the report still prints.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch os.Args[1] {
	case "run":
		return cmdRun(ctx, os.Args[2:])
	case "validate":
		return cmdValidate(ctx, os.Args[2:])
	case "provision":
		return cmdProvision(ctx, os.Args[2:])
	case "-h", "--help", "help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown subcommand %q", os.Args[1])
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `perf — transaction throughput and latency testing

Usage:
  perf run       -config <file> [-config <overlay>] [flags]
  perf validate  -config <file> [-config <overlay>]
  perf provision -count <n> [flags]

Commands:
  run        Execute a scenario and report throughput and latency.
  validate   Load a scenario, check it, and print the effective settings.
  provision  Create local development wallets and emit a scenario overlay.

Run "perf <command> -h" for the flags of each command.
`)
}

// configFlag collects repeatable -config values, which configa deep-merges in
// order. That is what keeps run shape in git and credentials out of it.
type configFlag []string

func (c *configFlag) String() string { return strings.Join(*c, ",") }

func (c *configFlag) Set(v string) error {
	*c = append(*c, v)
	return nil
}

func cmdRun(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	var files configFlag
	fs.Var(&files, "config", "scenario YAML file; repeat to layer overlays (later files win)")

	// Flag overrides for the knobs most likely to change between runs. Anything
	// left at its zero value keeps the scenario file's value.
	target := fs.String("target", "", "override target.grpc_address")
	rate := fs.Float64("rate", 0, "override run.arrival_rate (payments/sec, 0 keeps the scenario value)")
	duration := fs.Duration("duration", 0, "override run.duration")
	count := fs.Int("count", 0, "override run.count_per_sender")
	stopMode := fs.String("stop", "", "override run.stop (drain|count|duration)")
	pairing := fs.String("pairing", "", "override run.pairing (index|fan_in|round_robin|random)")
	metricsAddr := fs.String("metrics", "", "override metrics.listen, e.g. :9464")
	noSettlement := fs.Bool("no-settlement", false, "disable settlement tracking (measures RPC latency only)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("at least one -config file is required")
	}

	cfg, err := config.Load(ctx, files)
	if err != nil {
		return err
	}

	applyOverrides(cfg, overrides{
		target:       *target,
		rate:         *rate,
		duration:     *duration,
		count:        *count,
		stopMode:     *stopMode,
		pairing:      *pairing,
		metricsAddr:  *metricsAddr,
		noSettlement: *noSettlement,
	})

	r := runner.New(cfg, os.Stdout)

	var server *metrics.Server
	if cfg.Metrics.Listen != "" {
		server = metrics.Serve(cfg.Metrics.Listen, r.Collectors())
		fmt.Fprintf(os.Stdout, "metrics on http://localhost%s/metrics\n", cfg.Metrics.Listen)
	}

	runErr := r.Run(ctx)

	if server != nil {
		// Hold the endpoint open briefly so Prometheus can scrape the final values;
		// without this the last few seconds of a run are simply missing.
		if cfg.Metrics.LingerAfterRun > 0 && runErr == nil {
			fmt.Fprintf(os.Stdout, "\nholding metrics endpoint open for %s (ctrl-c to exit)\n", cfg.Metrics.LingerAfterRun)
			select {
			case <-time.After(cfg.Metrics.LingerAfterRun):
			case <-ctx.Done():
			}
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)

		select {
		case err := <-server.Err():
			if err != nil {
				fmt.Fprintf(os.Stderr, "metrics endpoint: %v\n", err)
			}
		default:
		}
	}

	return runErr
}

type overrides struct {
	target       string
	rate         float64
	duration     time.Duration
	count        int
	stopMode     string
	pairing      string
	metricsAddr  string
	noSettlement bool
}

func applyOverrides(cfg *config.Config, o overrides) {
	if o.target != "" {
		cfg.Target.GRPCAddress = o.target
	}
	if o.rate > 0 {
		cfg.Run.ArrivalRate = o.rate
	}
	if o.duration > 0 {
		cfg.Run.Duration = o.duration
	}
	if o.count > 0 {
		cfg.Run.CountPerSender = o.count
	}
	if o.stopMode != "" {
		cfg.Run.Stop = config.StopMode(o.stopMode)
	}
	if o.pairing != "" {
		cfg.Run.Pairing = config.Pairing(o.pairing)
	}
	if o.metricsAddr != "" {
		cfg.Metrics.Listen = o.metricsAddr
	}
	if o.noSettlement {
		cfg.Settlement.Track = false
	}
}

func cmdValidate(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	var files configFlag
	fs.Var(&files, "config", "scenario YAML file; repeat to layer overlays")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("at least one -config file is required")
	}

	cfg, err := config.Load(ctx, files)
	if err != nil {
		return err
	}

	fmt.Printf("scenario is valid\n\n")
	fmt.Printf("target        %s (kratos %s)\n", cfg.Target.GRPCAddress, cfg.Target.KratosURL)
	fmt.Printf("shape         %d senders → %d receivers, %s pairing\n", len(cfg.Senders), len(cfg.Receivers), cfg.Run.Pairing)
	fmt.Printf("stop          %s\n", cfg.Run.Stop)
	if cfg.Run.Duration > 0 {
		fmt.Printf("duration      %s\n", cfg.Run.Duration)
	}
	if cfg.Run.Stop == config.StopCount {
		fmt.Printf("count         %d per sender\n", cfg.Run.CountPerSender)
	}
	fmt.Printf("amount        %d minor units per payment\n", cfg.Run.Amount)
	if cfg.Run.ArrivalRate > 0 {
		fmt.Printf("arrival rate  %.1f/s (burst %d)\n", cfg.Run.ArrivalRate, cfg.Run.Burst)
	} else {
		fmt.Printf("arrival rate  unthrottled\n")
	}
	fmt.Printf("max in flight %d\n", cfg.Run.MaxInFlight)
	fmt.Printf("settlement    track=%t poll=%s timeout=%s workers=%d\n",
		cfg.Settlement.Track, cfg.Settlement.PollInterval, cfg.Settlement.Timeout, cfg.Settlement.Workers)
	fmt.Printf("connections   %d\n", cfg.Target.Connections)
	if cfg.Metrics.Listen != "" {
		fmt.Printf("metrics       %s\n", cfg.Metrics.Listen)
	}

	if !cfg.Settlement.Track {
		fmt.Printf("\nwarning: settlement tracking is off, so this run measures RPC latency only.\n")
		fmt.Printf("         CreatePayment returns before any money moves.\n")
	}

	return nil
}

func cmdProvision(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("provision", flag.ExitOnError)
	countries := fs.String("countries", "", "comma-separated country codes to provision, e.g. za,de,us")
	perCountry := fs.Int("per-country", 100, "wallets to create per country")
	targetMajor := fs.Int64("target", 5000, "target balance in major units for each wallet")
	prefix := fs.String("prefix", "perf", "prefix for generated emails, phone numbers and wallet addresses")
	password := fs.String("password", "", "password for created identities (default: a fixed local test password)")
	addressHost := fs.String("address-host", "https://local.ilp.link", "wallet address host")
	grpcAddr := fs.String("grpc-address", "localhost:8443", "backend gRPC address")
	kratosURL := fs.String("kratos", "http://localhost:4433", "public Kratos URL")
	dsn := fs.String("dsn", "postgres://postgres:postgres@localhost:5432/backend?sslmode=disable",
		"backend Postgres DSN, used to approve KYC; empty skips approval")
	out := fs.String("out", "", "write the scenario overlay here instead of stdout")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *countries == "" {
		return errors.New("-countries is required")
	}
	if *perCountry < 1 {
		return errors.New("-per-country must be at least 1")
	}
	if *targetMajor < 1 {
		return errors.New("-target must be at least 1")
	}

	fmt.Printf("provisioning %s against %s\n", *countries, *grpcAddr)
	fmt.Printf("this creates real records in the target environment — only run it against local development\n\n")

	wallets, err := provision.Run(ctx, provision.Options{
		Countries:   strings.Split(*countries, ","),
		PerCountry:  *perCountry,
		TargetMajor: int64(*targetMajor),
		Prefix:      *prefix,
		Password:    *password,
		AddressHost: *addressHost,
		GRPCAddress: *grpcAddr,
		KratosURL:   *kratosURL,
		BackendDSN:  *dsn,
	}, os.Stdout)
	if err != nil {
		return err
	}

	dest := os.Stdout
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			return fmt.Errorf("create %s: %w", *out, err)
		}
		defer f.Close()
		dest = f
	} else {
		fmt.Printf("\n── scenario overlay ──\n")
	}

	if err := provision.WriteScenarioOverlay(dest, wallets); err != nil {
		return err
	}

	if *out != "" {
		fmt.Printf("\nwrote %d wallets to %s\n", len(wallets), *out)
	}

	return nil
}
