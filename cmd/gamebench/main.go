// gamebench: game-OLTP benchmark for Azure MySQL (v1/v2 storage comparison).
//
// Subcommands:
//   schema   create tables
//   reset    truncate tables
//   load     seed dataset (parallel, random payloads)
//   env      print env snapshot (variables, sizes) as JSON
//   check    application invariants (G9) as JSON
//   knee     closed-loop concurrency staircase -> knee estimate JSON
//   run      open-loop (or closed) measurement -> run result JSON with gates
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/euson/azure-db-benchmark/internal/bench"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func logf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, time.Now().UTC().Format("2006-01-02T15:04:05Z")+" "+format+"\n", a...)
}

var driver = "mysql"

func mustDB(dsn string, maxOpen int) *sql.DB {
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
		if driver == "pgx" {
			dsn = os.Getenv("PG_DSN")
		}
	}
	if dsn == "" {
		logf("missing -dsn / MYSQL_DSN|PG_DSN")
		os.Exit(2)
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		logf("open: %v", err)
		os.Exit(2)
	}
	db.SetMaxOpenConns(maxOpen + 8)
	db.SetMaxIdleConns(maxOpen + 8)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(0)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		logf("ping: %v", err)
		os.Exit(2)
	}
	return db
}

func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		logf("marshal: %v", err)
		os.Exit(1)
	}
	if path == "" || path == "-" {
		os.Stdout.Write(b)
		os.Stdout.Write([]byte("\n"))
		return
	}
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		logf("write %s: %v", path, err)
		os.Exit(1)
	}
	logf("wrote %s", path)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gamebench <schema|reset|load|env|check|knee|run> [flags]")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	sub := os.Args[1]
	fs := flag.NewFlagSet(sub, flag.ExitOnError)
	dsn := fs.String("dsn", "", "MySQL DSN (or MYSQL_DSN) / PG URL (or PG_DSN)")
	drv := fs.String("driver", "mysql", "mysql|pgx")
	schema := fs.String("schema", "benchmark", "database name (for size queries)")
	out := fs.String("out", "-", "output JSON path")
	// load flags
	accounts := fs.Int64("accounts", 5_000_000, "number of accounts")
	slots := fs.Int("slots", 20, "inventory slots per account")
	profileBytes := fs.Int("profile-bytes", 1024, "random profile payload size")
	attrsBytes := fs.Int("attrs-bytes", 256, "random inventory attrs size")
	guilds := fs.Int("guilds", 100_000, "guilds")
	loadWorkers := fs.Int("load-workers", 16, "parallel loader connections")
	batch := fs.Int("batch", 500, "accounts per loader transaction")
	seed := fs.Uint64("seed", 20260818, "seed")
	// run flags
	mode := fs.String("mode", "open", "open|closed")
	scenario := fs.String("scenario", "S1", "S1|S2|S3|S4|read")
	rate := fs.Float64("rate", 1000, "open-loop target arrivals/s")
	workers := fs.Int("workers", 256, "workers = dedicated connections = max in-flight")
	warmup := fs.Int("warmup", 120, "min warmup seconds")
	warmupMax := fs.Int("warmup-max", 600, "max warmup seconds")
	steadyCV := fs.Float64("steady-cv", 5.0, "steady-state CV%% threshold over last 120s")
	measure := fs.Int("measure", 600, "measurement seconds")
	queueSec := fs.Float64("queue-sec", 10, "admission queue depth (seconds of arrivals)")
	stmtTimeout := fs.Int("stmt-timeout-ms", 5000, "per-operation deadline")
	hotKeys := fs.Int64("hot-keys", 50_000, "hot set size (accounts)")
	hotProb := fs.Float64("hot-prob", 0.2, "probability an op targets the hot set")
	nic := fs.Float64("nic-bps", 0, "NIC limit bytes/s for G3 (0=unknown)")
	label := fs.String("label", "", "arm label (v1/v2)")
	burstAt := fs.Int("burst-at", 0, "S3: burst start second (0=off)")
	burstSec := fs.Int("burst-sec", 0, "S3: burst length seconds")
	burstFactor := fs.Float64("burst-factor", 1, "S3: rate multiplier during burst")
	// knee flags
	steps := fs.String("steps", "16,32,64,128,256,512", "closed-loop concurrency steps")
	stepSec := fs.Int("step-sec", 120, "seconds per step (after 30s warmup)")
	fs.Parse(os.Args[2:])
	driver = *drv
	isPG := driver == "pgx"
	createSchema, reset, load, captureEnv, check, scenarioMix := bench.CreateSchema, bench.Reset, bench.Load, bench.CaptureEnv, bench.CheckInvariants, bench.ScenarioMix
	var fetch bench.StatusFetch
	if isPG {
		createSchema, reset, load, captureEnv, check, scenarioMix, fetch = bench.PGCreateSchema, bench.PGReset, bench.PGLoad, bench.PGCaptureEnv, bench.PGCheckInvariants, bench.PGScenarioMix, bench.PGFetchStatus
	}

	switch sub {
	case "schema":
		db := mustDB(*dsn, 4)
		if err := createSchema(ctx, db); err != nil {
			logf("schema: %v", err)
			os.Exit(1)
		}
		logf("schema ok")
	case "reset":
		db := mustDB(*dsn, 4)
		if err := reset(ctx, db); err != nil {
			logf("reset: %v", err)
			os.Exit(1)
		}
		logf("reset ok")
	case "load":
		db := mustDB(*dsn, *loadWorkers)
		cfg := bench.LoadConfig{Accounts: *accounts, Slots: *slots, ProfileBytes: *profileBytes, AttrsBytes: *attrsBytes, Guilds: *guilds, Workers: *loadWorkers, Batch: *batch, Seed: *seed, InitialBalance: 1_000_000}
		if err := load(ctx, db, cfg, logf); err != nil {
			logf("load: %v", err)
			os.Exit(1)
		}
		env, _ := captureEnv(ctx, db, *schema)
		writeJSON(*out, env)
	case "env":
		db := mustDB(*dsn, 4)
		env, err := captureEnv(ctx, db, *schema)
		if err != nil {
			logf("env: %v", err)
		}
		writeJSON(*out, env)
	case "check":
		db := mustDB(*dsn, 4)
		rep, err := check(ctx, db)
		if err != nil {
			logf("check: %v", err)
			os.Exit(1)
		}
		writeJSON(*out, rep)
	case "knee":
		db := mustDB(*dsn, 1024)
		mix, err := scenarioMix(*scenario)
		if err != nil {
			logf("%v", err)
			os.Exit(2)
		}
		env, _ := captureEnv(ctx, db, *schema)
		var ns []int
		for _, s := range splitInts(*steps) {
			ns = append(ns, s)
		}
		type stepRes struct {
			Workers    int     `json:"workers"`
			SuccessTPS float64 `json:"success_tps"`
			P50Us      int64   `json:"p50_us"`
			P99Us      int64   `json:"p99_us"`
			ErrorRate  float64 `json:"error_rate"`
			ReadIOPS   float64 `json:"read_iops"`
			WriteIOPS  float64 `json:"write_iops"`
			HostCPU    float64 `json:"host_cpu_max"`
		}
		var results []stepRes
		bestTPS, kneeWorkers := 0.0, 0
		for i, n := range ns {
			cfg := bench.RunConfig{Mode: "closed", Scenario: *scenario, Workers: n, WarmupSec: 30, WarmupMaxSec: 30, SteadyCVPct: 100, MeasureSec: *stepSec,
				StmtTimeoutMs: *stmtTimeout, Accounts: *accounts, Slots: *slots, HotKeys: *hotKeys, HotProb: *hotProb, Seed: *seed + uint64(i), NICLimitBps: *nic, Label: fmt.Sprintf("%s-knee-%d", *label, n)}
			r := &bench.Runner{DB: db, Cfg: cfg, Mix: mix, Log: logf, Env: env, Fetch: fetch}
			res, err := r.Run(ctx)
			if err != nil {
				logf("knee step %d: %v", n, err)
				os.Exit(1)
			}
			sr := stepRes{Workers: n, SuccessTPS: res.SuccessTPS, P50Us: res.Overall.Latency.P50, P99Us: res.Overall.Latency.P99, ErrorRate: res.ErrorRate, HostCPU: res.Host.CPUMax}
			if res.StatusDelta != nil {
				sr.ReadIOPS, sr.WriteIOPS = res.StatusDelta.ReadIOPS, res.StatusDelta.WriteIOPS
			}
			results = append(results, sr)
			logf("knee step workers=%d tps=%.0f p50=%dus p99=%dus err=%.3f%% read_iops=%.0f", n, sr.SuccessTPS, sr.P50Us, sr.P99Us, 100*sr.ErrorRate, sr.ReadIOPS)
			// knee: throughput gain <10% vs previous step, or p99 doubled, or errors
			if i > 0 {
				prev := results[i-1]
				gain := (sr.SuccessTPS - prev.SuccessTPS) / prev.SuccessTPS
				if gain < 0.10 || sr.P99Us > 2*prev.P99Us || sr.ErrorRate > 0.01 {
					if kneeWorkers == 0 {
						kneeWorkers = prev.Workers
						bestTPS = prev.SuccessTPS
						if sr.SuccessTPS > bestTPS && sr.ErrorRate <= 0.01 {
							bestTPS = sr.SuccessTPS
						}
					}
					logf("knee detected at workers=%d (sustainable ~%.0f tps); stopping staircase", kneeWorkers, bestTPS)
					break
				}
			}
			if sr.SuccessTPS > bestTPS {
				bestTPS = sr.SuccessTPS
			}
		}
		if kneeWorkers == 0 && len(results) > 0 {
			kneeWorkers = results[len(results)-1].Workers
		}
		writeJSON(*out, map[string]any{"scenario": *scenario, "label": *label, "steps": results, "knee_workers": kneeWorkers, "knee_tps": bestTPS,
			"recommended_rate_65pct": bestTPS * 0.65, "env": env, "captured_at": time.Now().UTC()})
	case "run":
		db := mustDB(*dsn, *workers)
		mix, err := scenarioMix(*scenario)
		if err != nil {
			logf("%v", err)
			os.Exit(2)
		}
		env, err := captureEnv(ctx, db, *schema)
		if err != nil {
			logf("env warning: %v", err)
		}
		hk, hp := *hotKeys, *hotProb
		if *scenario == "S4" || *scenario == "s4" || *scenario == "hotspot" {
			hk, hp = 1000, 0.5
		}
		cfg := bench.RunConfig{Mode: *mode, Scenario: *scenario, TargetRate: *rate, Workers: *workers, WarmupSec: *warmup, WarmupMaxSec: *warmupMax, SteadyCVPct: *steadyCV,
			MeasureSec: *measure, QueueSec: *queueSec, StmtTimeoutMs: *stmtTimeout, Accounts: *accounts, Slots: *slots, HotKeys: hk, HotProb: hp, Seed: *seed,
			NICLimitBps: *nic, Label: *label, BurstAtSec: *burstAt, BurstSec: *burstSec, BurstFactor: *burstFactor}
		r := &bench.Runner{DB: db, Cfg: cfg, Mix: mix, Log: logf, Env: env, Fetch: fetch}
		res, err := r.Run(ctx)
		if err != nil {
			logf("run: %v", err)
			os.Exit(1)
		}
		logf("[%s] done: tps=%.1f p50=%dus p95=%dus p99=%dus p999=%dus err=%.4f%% timeout=%.4f%% drop=%.4f%% queue_p99=%dus gates_passed=%v",
			cfg.Label, res.SuccessTPS, res.Overall.Latency.P50, res.Overall.Latency.P95, res.Overall.Latency.P99, res.Overall.Latency.P999,
			100*res.ErrorRate, 100*res.TimeoutRate, 100*res.DropRate, res.QueueDelay.P99, res.GatesPassed)
		for k, g := range res.Gates {
			logf("  gate %-26s %-5v %s (%s)", k, g.Pass, g.Value, g.Rule)
		}
		if res.StatusDelta != nil {
			d := res.StatusDelta
			logf("  server: read_iops=%.0f write_iops=%.0f read=%.1fMB/s write=%.1fMB/s hit=%.4f threads_running_max=%d", d.ReadIOPS, d.WriteIOPS, d.ReadMBps, d.WriteMBps, d.BufferPoolHitRatio, d.MaxThreadsRunning)
		}
		writeJSON(*out, res)
		if !res.GatesPassed {
			os.Exit(3)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown subcommand", sub)
		os.Exit(2)
	}
}

func splitInts(s string) []int {
	var out []int
	cur := 0
	has := false
	for _, ch := range s + "," {
		if ch >= '0' && ch <= '9' {
			cur = cur*10 + int(ch-'0')
			has = true
		} else {
			if has {
				out = append(out, cur)
			}
			cur, has = 0, false
		}
	}
	return out
}
