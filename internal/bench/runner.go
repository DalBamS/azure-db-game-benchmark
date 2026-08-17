package bench

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

type RunConfig struct {
	Mode           string  `json:"mode"` // "open" or "closed"
	Scenario       string  `json:"scenario"`
	TargetRate     float64 `json:"target_rate"` // open-loop arrivals/s
	Workers        int     `json:"workers"`     // = dedicated connections = max in-flight
	WarmupSec      int     `json:"warmup_sec"`
	WarmupMaxSec   int     `json:"warmup_max_sec"`
	SteadyCVPct    float64 `json:"steady_cv_pct"`
	MeasureSec     int     `json:"measure_sec"`
	QueueSec       float64 `json:"queue_sec"`       // open-loop admission queue depth in seconds of arrivals
	StmtTimeoutMs  int     `json:"stmt_timeout_ms"` // per-operation deadline
	Accounts       int64   `json:"accounts"`
	Slots          int     `json:"slots"`
	HotKeys        int64   `json:"hot_keys"`
	HotProb        float64 `json:"hot_prob"`
	Seed           uint64  `json:"seed"`
	StatusEverySec int     `json:"status_every_sec"`
	NICLimitBps    float64 `json:"nic_limit_bps"`
	Label          string  `json:"label"` // arm label e.g. v1 / v2
	// burst (S3): after BurstAtSec seconds of measurement, multiply rate by BurstFactor for BurstSec seconds
	BurstAtSec  int     `json:"burst_at_sec"`
	BurstSec    int     `json:"burst_sec"`
	BurstFactor float64 `json:"burst_factor"`
}

type OpResult struct {
	Attempts uint64    `json:"attempts"`
	Success  uint64    `json:"success"`
	Errors   uint64    `json:"errors"`
	Timeouts uint64    `json:"timeouts"`
	Latency  Quantiles `json:"latency"`
}

type SecondSample struct {
	T        time.Time `json:"t"`
	Success  uint64    `json:"success"`
	Errors   uint64    `json:"errors"`
	Arrivals uint64    `json:"arrivals"`
	P99Us    int64     `json:"p99_us"`
	InFlight int32     `json:"in_flight"`
	QueueLen int       `json:"queue_len"`
}

type RunResult struct {
	Config        RunConfig            `json:"config"`
	Mix           map[string]int       `json:"mix"`
	KeyModel      map[string]float64   `json:"key_model"`
	StartedAt     time.Time            `json:"started_at"`
	WarmupFrom    time.Time            `json:"warmup_from"`
	MeasureFrom   time.Time            `json:"measure_from"`
	MeasureTo     time.Time            `json:"measure_to"`
	SteadyState   SteadyInfo           `json:"steady_state"`
	Overall       OpResult             `json:"overall"`
	QueueDelay    Quantiles            `json:"queue_delay"`
	ServiceTime   Quantiles            `json:"service_time"`
	PerOp         map[string]*OpResult `json:"per_op"`
	Scheduled     uint64               `json:"scheduled"`
	Dropped       uint64               `json:"dropped"`
	SuccessTPS    float64              `json:"success_tps"`
	ErrorRate     float64              `json:"error_rate"`
	TimeoutRate   float64              `json:"timeout_rate"`
	DropRate      float64              `json:"drop_rate"`
	ErrorClasses  map[string]uint64    `json:"error_classes"`
	Seconds       []SecondSample       `json:"per_second"`
	Status        []StatusSample       `json:"server_status"`
	StatusDelta   *StatusDelta         `json:"server_status_delta"`
	StatusErrors  int                  `json:"server_status_errors"`
	Host          HostSummary          `json:"host"`
	HostSamples   []HostSample         `json:"host_samples"`
	Env           *EnvSnapshot         `json:"env"`
	Histogram     string               `json:"hdr_overall_b64,omitempty"`
	MaxInFlight   int32                `json:"max_in_flight"`
	Gates         map[string]GateResult `json:"gates"`
	GatesPassed   bool                 `json:"gates_passed"`
}

type SteadyInfo struct {
	Reached      bool    `json:"reached"`
	WarmupSecUsed int    `json:"warmup_sec_used"`
	CVPct        float64 `json:"cv_pct_last_120s"`
}

type opStats struct {
	attempts, success, errors, timeouts atomic.Uint64
	lat                                 *LatencyHist
}

// Runner executes one arm run.
type Runner struct {
	DB   *sql.DB
	Cfg  RunConfig
	Mix  *Mix
	Log  func(string, ...any)
	Env  *EnvSnapshot
	Fetch StatusFetch // nil = MySQL SHOW GLOBAL STATUS
}

func classify(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return fmt.Sprintf("mysql_%d", me.Number)
	}
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return "pg_" + pe.Code
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "timeout"), strings.Contains(s, "i/o timeout"):
		return "timeout"
	case strings.Contains(s, "connection refused"), strings.Contains(s, "broken pipe"), strings.Contains(s, "EOF"), strings.Contains(s, "bad connection"):
		return "conn"
	}
	return "other"
}

func (r *Runner) Run(ctx context.Context) (*RunResult, error) {
	cfg := r.Cfg
	if cfg.Workers <= 0 {
		cfg.Workers = 256
	}
	if cfg.StatusEverySec <= 0 {
		cfg.StatusEverySec = 5
	}
	res := &RunResult{Config: cfg, Mix: r.Mix.Describe(), PerOp: map[string]*OpResult{}, ErrorClasses: map[string]uint64{}, Gates: map[string]GateResult{}}
	res.StartedAt = time.Now().UTC()

	// dedicated connections
	conns := make([]*sql.Conn, cfg.Workers)
	for i := range conns {
		c, err := r.DB.Conn(ctx)
		if err != nil {
			return nil, fmt.Errorf("open conn %d: %w", i, err)
		}
		conns[i] = c
	}
	defer func() {
		for _, c := range conns {
			if c != nil {
				c.Close()
			}
		}
	}()

	stats := map[string]*opStats{}
	for _, op := range r.Mix.Ops {
		stats[op.Name] = &opStats{lat: NewLatencyHist()}
	}
	overall := &opStats{lat: NewLatencyHist()}
	queueHist := NewLatencyHist()
	serviceHist := NewLatencyHist()
	errClasses := sync.Map{}

	var measuring atomic.Bool
	var inFlight atomic.Int32
	var maxInFlight atomic.Int32
	var secSuccess, secErrors, secArrivals atomic.Uint64
	secHist := NewLatencyHist()

	// admission queue
	qcap := int(cfg.TargetRate * cfg.QueueSec)
	if qcap < cfg.Workers*2 {
		qcap = cfg.Workers * 2
	}
	jobs := make(chan time.Time, qcap)

	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	kp := NewKeyPicker(cfg.Accounts, cfg.HotKeys, cfg.HotProb, cfg.Seed)
	top1, hot := kp.TopShare()
	res.KeyModel = map[string]float64{"top1_share": top1, "hotset_share": hot, "hot_keys": float64(cfg.HotKeys), "hot_prob": cfg.HotProb, "accounts": float64(cfg.Accounts)}

	execOne := func(w *Worker, conn *sql.Conn, scheduledAt time.Time) {
		start := time.Now()
		qd := start.Sub(scheduledAt)
		op := r.Mix.Pick(w.Rng)
		st := stats[op.Name]
		inF := inFlight.Add(1)
		for {
			old := maxInFlight.Load()
			if inF <= old || maxInFlight.CompareAndSwap(old, inF) {
				break
			}
		}
		opCtx, cancel := context.WithTimeout(runCtx, time.Duration(cfg.StmtTimeoutMs)*time.Millisecond)
		err := op.Run(opCtx, conn, w)
		cancel()
		inFlight.Add(-1)
		end := time.Now()
		if runCtx.Err() != nil {
			return // run ended while op in flight: don't count
		}
		lat := end.Sub(scheduledAt).Microseconds()
		svc := end.Sub(start).Microseconds()
		st.attempts.Add(1)
		overall.attempts.Add(1)
		if err == nil {
			st.success.Add(1)
			overall.success.Add(1)
			st.lat.Record(lat)
			overall.lat.Record(lat)
			queueHist.Record(qd.Microseconds())
			serviceHist.Record(svc)
			secSuccess.Add(1)
			secHist.Record(lat)
		} else {
			st.errors.Add(1)
			overall.errors.Add(1)
			secErrors.Add(1)
			cls := classify(err)
			if cls == "timeout" {
				st.timeouts.Add(1)
				overall.timeouts.Add(1)
			}
			v, _ := errClasses.LoadOrStore(cls, new(atomic.Uint64))
			v.(*atomic.Uint64).Add(1)
			// a broken connection is replaced so the worker keeps going
			if cls == "conn" {
				conn.Close()
				if nc, e := r.DB.Conn(runCtx); e == nil {
					conns[w.ID] = nc
				}
			}
		}
	}

	runNonce := uint64(time.Now().UnixNano())
	var wg sync.WaitGroup
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// per-run entropy: request ids (uuids) must not repeat across runs, so mix wall-clock nanos into the seed
			w := &Worker{ID: id, Rng: rand.New(rand.NewPCG(cfg.Seed+uint64(id)*7919+runNonce, cfg.Seed^uint64(id)^runNonce)), Keys: NewKeyPicker(cfg.Accounts, cfg.HotKeys, cfg.HotProb, cfg.Seed+uint64(id)*104729), Slots: cfg.Slots}
			if cfg.Mode == "closed" {
				for runCtx.Err() == nil {
					execOne(w, conns[w.ID], time.Now())
				}
				return
			}
			for sched := range jobs {
				if runCtx.Err() != nil {
					return
				}
				execOne(w, conns[w.ID], sched)
			}
		}(i)
	}

	// samplers
	statusSampler := NewStatusSampler(r.DB, r.Fetch, time.Duration(cfg.StatusEverySec)*time.Second)
	statusSampler.Start(ctx)
	host := NewHostSampler(5 * time.Second)
	host.Start()

	// per-second recorder + steady-state detector
	var secMu sync.Mutex
	var seconds []SecondSample
	warmupSeconds := []uint64{}
	secDone := make(chan struct{})
	go func() {
		defer close(secDone)
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case now := <-t.C:
				s := SecondSample{T: now.UTC(), Success: secSuccess.Swap(0), Errors: secErrors.Swap(0), Arrivals: secArrivals.Swap(0), InFlight: inFlight.Load(), QueueLen: len(jobs)}
				s.P99Us = secHist.Snapshot().P99
				secHist.Reset()
				secMu.Lock()
				if measuring.Load() {
					seconds = append(seconds, s)
				} else {
					warmupSeconds = append(warmupSeconds, s.Success)
				}
				secMu.Unlock()
			}
		}
	}()

	// scheduler (open loop)
	var scheduled, dropped atomic.Uint64
	rateNow := func(elapsedMeasure float64) float64 {
		if cfg.BurstFactor > 1 && elapsedMeasure >= float64(cfg.BurstAtSec) && elapsedMeasure < float64(cfg.BurstAtSec+cfg.BurstSec) {
			return cfg.TargetRate * cfg.BurstFactor
		}
		return cfg.TargetRate
	}
	schedDone := make(chan struct{})
	var measureStart atomic.Int64 // unix nanos, 0 = not measuring
	if cfg.Mode == "open" {
		go func() {
			defer close(schedDone)
			defer close(jobs)
			next := time.Now()
			for runCtx.Err() == nil {
				em := -1.0
				if ms := measureStart.Load(); ms > 0 {
					em = time.Since(time.Unix(0, ms)).Seconds()
				}
				interval := time.Duration(float64(time.Second) / rateNow(em))
				now := time.Now()
				if next.After(now) {
					time.Sleep(next.Sub(now))
					if runCtx.Err() != nil {
						return
					}
				}
				scheduled.Add(1)
				secArrivals.Add(1)
				select {
				case jobs <- next:
				default:
					dropped.Add(1)
				}
				next = next.Add(interval)
				// if we fell far behind (e.g. GC pause), don't try to catch up more than 1s
				if lag := time.Since(next); lag > time.Second {
					next = time.Now()
				}
			}
		}()
	} else {
		close(schedDone)
	}

	// warmup with steady-state gate
	res.WarmupFrom = time.Now().UTC()
	r.Log("[%s] warmup start (min %ds, max %ds, CV<%.1f%%)", cfg.Label, cfg.WarmupSec, cfg.WarmupMaxSec, cfg.SteadyCVPct)
	warmStart := time.Now()
	steady := SteadyInfo{}
	for {
		elapsed := int(time.Since(warmStart).Seconds())
		if elapsed >= cfg.WarmupSec {
			secMu.Lock()
			cv := cvLast(warmupSeconds, 120)
			secMu.Unlock()
			steady.CVPct = cv
			if cv >= 0 && cv <= cfg.SteadyCVPct {
				steady.Reached = true
				steady.WarmupSecUsed = elapsed
				break
			}
			if elapsed >= cfg.WarmupMaxSec {
				steady.WarmupSecUsed = elapsed
				r.Log("[%s] WARNING steady state not reached (CV=%.2f%%), starting measurement anyway (G7 will fail)", cfg.Label, cv)
				break
			}
			r.Log("[%s] warmup %ds: CV=%.2f%% > %.1f%%, extending", cfg.Label, elapsed, cv, cfg.SteadyCVPct)
		}
		select {
		case <-ctx.Done():
			cancelRun()
			return nil, ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	res.SteadyState = steady

	// reset stats and start measuring
	for _, s := range stats {
		s.attempts.Store(0); s.success.Store(0); s.errors.Store(0); s.timeouts.Store(0); s.lat.Reset()
	}
	overall.attempts.Store(0); overall.success.Store(0); overall.errors.Store(0); overall.timeouts.Store(0); overall.lat.Reset()
	queueHist.Reset(); serviceHist.Reset()
	scheduled.Store(0); dropped.Store(0)
	errClasses.Range(func(k, v any) bool { errClasses.Delete(k); return true })
	res.MeasureFrom = time.Now().UTC()
	measureStart.Store(res.MeasureFrom.UnixNano())
	measuring.Store(true)
	r.Log("[%s] measuring %ds at %.0f/s (%s, workers=%d)", cfg.Label, cfg.MeasureSec, cfg.TargetRate, cfg.Mode, cfg.Workers)
	select {
	case <-ctx.Done():
		cancelRun()
		return nil, ctx.Err()
	case <-time.After(time.Duration(cfg.MeasureSec) * time.Second):
	}
	res.MeasureTo = time.Now().UTC()
	measuring.Store(false)
	cancelRun()
	<-schedDone
	wg.Wait()
	<-secDone
	res.Status = statusSampler.Stop()
	res.StatusErrors = statusSampler.Errors()
	res.HostSamples = host.Stop()

	// aggregate
	res.Scheduled, res.Dropped = scheduled.Load(), dropped.Load()
	res.Overall = OpResult{Attempts: overall.attempts.Load(), Success: overall.success.Load(), Errors: overall.errors.Load(), Timeouts: overall.timeouts.Load(), Latency: overall.lat.Snapshot()}
	res.QueueDelay = queueHist.Snapshot()
	res.ServiceTime = serviceHist.Snapshot()
	for name, s := range stats {
		res.PerOp[name] = &OpResult{Attempts: s.attempts.Load(), Success: s.success.Load(), Errors: s.errors.Load(), Timeouts: s.timeouts.Load(), Latency: s.lat.Snapshot()}
	}
	errClasses.Range(func(k, v any) bool { res.ErrorClasses[k.(string)] = v.(*atomic.Uint64).Load(); return true })
	dur := res.MeasureTo.Sub(res.MeasureFrom).Seconds()
	res.SuccessTPS = float64(res.Overall.Success) / dur
	if res.Overall.Attempts > 0 {
		res.ErrorRate = float64(res.Overall.Errors) / float64(res.Overall.Attempts)
		res.TimeoutRate = float64(res.Overall.Timeouts) / float64(res.Overall.Attempts)
	}
	if res.Scheduled > 0 {
		res.DropRate = float64(res.Dropped) / float64(res.Scheduled)
	}
	secMu.Lock()
	res.Seconds = seconds
	secMu.Unlock()
	res.MaxInFlight = maxInFlight.Load()
	if d, err := ComputeDelta(res.Status, res.MeasureFrom, res.MeasureTo); err == nil {
		res.StatusDelta = d
	} else {
		r.Log("[%s] WARNING status delta: %v", cfg.Label, err)
	}
	res.Host = SummarizeHost(res.HostSamples, res.MeasureFrom, res.MeasureTo, cfg.NICLimitBps)
	res.Env = r.Env
	res.Histogram = encodeHist(overall.lat)
	res.Gates, res.GatesPassed = EvaluateGates(res)
	return res, nil
}

func cvLast(vals []uint64, n int) float64 {
	if len(vals) < n {
		if len(vals) < 30 {
			return -1
		}
		n = len(vals)
	}
	seg := vals[len(vals)-n:]
	sum := 0.0
	for _, v := range seg {
		sum += float64(v)
	}
	mean := sum / float64(len(seg))
	if mean == 0 {
		return -1
	}
	ss := 0.0
	for _, v := range seg {
		d := float64(v) - mean
		ss += d * d
	}
	return 100 * math.Sqrt(ss/float64(len(seg)-1)) / mean
}

// SortedOps returns op names sorted for stable output.
func SortedOps(m map[string]*OpResult) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
