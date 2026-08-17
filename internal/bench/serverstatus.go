package bench

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Counters sampled from SHOW GLOBAL STATUS. Deltas over the measurement window give
// server-side read/write IOPS and buffer-pool hit ratio independent of any external
// monitoring stack (gate G2/G6 evidence lives inside the run result).
var statusVars = []string{
	"Innodb_data_reads", "Innodb_data_writes", "Innodb_data_read", "Innodb_data_written",
	"Innodb_buffer_pool_reads", "Innodb_buffer_pool_read_requests", "Innodb_buffer_pool_pages_dirty",
	"Innodb_buffer_pool_wait_free", "Innodb_log_waits", "Innodb_os_log_written", "Innodb_row_lock_waits",
	"Innodb_row_lock_time", "Innodb_rows_read", "Innodb_rows_inserted", "Innodb_rows_updated",
	"Threads_connected", "Threads_running", "Questions", "Com_commit", "Com_rollback",
	"Innodb_dblwr_writes", "Innodb_dblwr_pages_written",
}

type StatusSample struct {
	T      time.Time         `json:"t"`
	Values map[string]uint64 `json:"v"`
}

type StatusSampler struct {
	db       *sql.DB
	interval time.Duration
	mu       sync.Mutex
	samples  []StatusSample
	errs     int
	stop     chan struct{}
	done     chan struct{}
}

func NewStatusSampler(db *sql.DB, interval time.Duration) *StatusSampler {
	return &StatusSampler{db: db, interval: interval, stop: make(chan struct{}), done: make(chan struct{})}
}

func FetchStatus(ctx context.Context, db *sql.DB) (map[string]uint64, error) {
	q := "SHOW GLOBAL STATUS WHERE Variable_name IN ("
	args := make([]any, len(statusVars))
	for i, v := range statusVars {
		if i > 0 {
			q += ","
		}
		q += "?"
		args[i] = v
	}
	q += ")"
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]uint64{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		n, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			out[k] = n
		}
	}
	return out, rows.Err()
}

func (s *StatusSampler) Start(ctx context.Context) {
	go func() {
		defer close(s.done)
		t := time.NewTicker(s.interval)
		defer t.Stop()
		s.sampleOnce(ctx)
		for {
			select {
			case <-s.stop:
				s.sampleOnce(ctx)
				return
			case <-ctx.Done():
				return
			case <-t.C:
				s.sampleOnce(ctx)
			}
		}
	}()
}

func (s *StatusSampler) sampleOnce(ctx context.Context) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	v, err := FetchStatus(c, s.db)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.errs++
		return
	}
	s.samples = append(s.samples, StatusSample{T: time.Now().UTC(), Values: v})
}

func (s *StatusSampler) Stop() []StatusSample {
	close(s.stop)
	<-s.done
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.samples
}

func (s *StatusSampler) Errors() int { s.mu.Lock(); defer s.mu.Unlock(); return s.errs }

// WindowDelta computes per-second rates over [from,to] using the nearest samples.
type StatusDelta struct {
	From, To            time.Time
	Seconds             float64
	ReadIOPS            float64 `json:"read_iops"`
	WriteIOPS           float64 `json:"write_iops"`
	ReadMBps            float64 `json:"read_mbps"`
	WriteMBps           float64 `json:"write_mbps"`
	BufferPoolHitRatio  float64 `json:"buffer_pool_hit_ratio"`
	BufferPoolReadsRate float64 `json:"buffer_pool_physical_reads_per_s"`
	LogWaitsRate        float64 `json:"log_waits_per_s"`
	RowLockWaitsRate    float64 `json:"row_lock_waits_per_s"`
	CommitsRate         float64 `json:"commits_per_s"`
	QuestionsRate       float64 `json:"questions_per_s"`
	MaxThreadsRunning   uint64  `json:"max_threads_running"`
	MaxThreadsConnected uint64  `json:"max_threads_connected"`
	Samples             int     `json:"samples_in_window"`
}

func ComputeDelta(samples []StatusSample, from, to time.Time) (*StatusDelta, error) {
	var first, last *StatusSample
	n := 0
	var maxRun, maxConn uint64
	for i := range samples {
		s := &samples[i]
		if s.T.Before(from.Add(-3*time.Second)) || s.T.After(to.Add(3*time.Second)) {
			continue
		}
		if first == nil {
			first = s
		}
		last = s
		n++
		if v := s.Values["Threads_running"]; v > maxRun {
			maxRun = v
		}
		if v := s.Values["Threads_connected"]; v > maxConn {
			maxConn = v
		}
	}
	if first == nil || last == nil || first == last {
		return nil, fmt.Errorf("not enough status samples in window (%d)", n)
	}
	sec := last.T.Sub(first.T).Seconds()
	d := func(k string) float64 { return float64(last.Values[k]-first.Values[k]) }
	bpReq, bpReads := d("Innodb_buffer_pool_read_requests"), d("Innodb_buffer_pool_reads")
	hit := 1.0
	if bpReq > 0 {
		hit = 1 - bpReads/bpReq
	}
	return &StatusDelta{
		From: first.T, To: last.T, Seconds: sec, Samples: n,
		ReadIOPS: d("Innodb_data_reads") / sec, WriteIOPS: d("Innodb_data_writes") / sec,
		ReadMBps: d("Innodb_data_read") / sec / 1048576, WriteMBps: d("Innodb_data_written") / sec / 1048576,
		BufferPoolHitRatio: hit, BufferPoolReadsRate: bpReads / sec,
		LogWaitsRate: d("Innodb_log_waits") / sec, RowLockWaitsRate: d("Innodb_row_lock_waits") / sec,
		CommitsRate: d("Com_commit") / sec, QuestionsRate: d("Questions") / sec,
		MaxThreadsRunning: maxRun, MaxThreadsConnected: maxConn,
	}, nil
}

// EnvSnapshot captures server variables & sizes relevant to the confound table.
type EnvSnapshot struct {
	Variables map[string]string `json:"variables"`
	TableGiB  map[string]float64 `json:"table_gib"`
	TotalGiB  float64            `json:"total_gib"`
	Rows      map[string]int64   `json:"approx_rows"`
	Hostname  string             `json:"hostname"`
	CapturedAt time.Time         `json:"captured_at"`
}

var envVars = []string{"version", "innodb_buffer_pool_size", "innodb_doublewrite", "innodb_flush_log_at_trx_commit",
	"sync_binlog", "innodb_redo_log_capacity", "innodb_io_capacity", "innodb_io_capacity_max", "innodb_flush_method",
	"innodb_page_size", "max_connections", "binlog_format", "log_bin", "innodb_flush_neighbors", "innodb_read_io_threads",
	"innodb_write_io_threads", "innodb_lru_scan_depth", "innodb_adaptive_hash_index", "require_secure_transport", "hostname"}

func CaptureEnv(ctx context.Context, db *sql.DB, schema string) (*EnvSnapshot, error) {
	e := &EnvSnapshot{Variables: map[string]string{}, TableGiB: map[string]float64{}, Rows: map[string]int64{}, CapturedAt: time.Now().UTC()}
	for _, v := range envVars {
		var name, val string
		if err := db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE ?", v).Scan(&name, &val); err == nil {
			e.Variables[name] = val
		}
	}
	e.Hostname = e.Variables["hostname"]
	rows, err := db.QueryContext(ctx, `SELECT table_name, (data_length+index_length)/1073741824.0, table_rows
		FROM information_schema.tables WHERE table_schema=?`, schema)
	if err != nil {
		return e, err
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		var gib float64
		var n int64
		if err := rows.Scan(&t, &gib, &n); err != nil {
			return e, err
		}
		e.TableGiB[t] = gib
		e.Rows[t] = n
		e.TotalGiB += gib
	}
	return e, rows.Err()
}
