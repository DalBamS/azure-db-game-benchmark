package bench

import (
	"sync"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// LatencyHist is a thread-safe HDR histogram in microseconds (1us .. 600s, 3 sig figs).
type LatencyHist struct {
	mu sync.Mutex
	h  *hdrhistogram.Histogram
}

func NewLatencyHist() *LatencyHist {
	return &LatencyHist{h: hdrhistogram.New(1, 600_000_000, 3)}
}

func (l *LatencyHist) Record(us int64) {
	if us < 1 {
		us = 1
	}
	l.mu.Lock()
	_ = l.h.RecordValue(us)
	l.mu.Unlock()
}

func (l *LatencyHist) Reset() {
	l.mu.Lock()
	l.h.Reset()
	l.mu.Unlock()
}

type Quantiles struct {
	Count int64   `json:"count"`
	Mean  float64 `json:"mean_us"`
	P50   int64   `json:"p50_us"`
	P90   int64   `json:"p90_us"`
	P95   int64   `json:"p95_us"`
	P99   int64   `json:"p99_us"`
	P999  int64   `json:"p999_us"`
	P9999 int64   `json:"p9999_us"`
	Max   int64   `json:"max_us"`
}

func (l *LatencyHist) Snapshot() Quantiles {
	l.mu.Lock()
	defer l.mu.Unlock()
	h := l.h
	return Quantiles{
		Count: h.TotalCount(), Mean: h.Mean(),
		P50: h.ValueAtQuantile(50), P90: h.ValueAtQuantile(90), P95: h.ValueAtQuantile(95),
		P99: h.ValueAtQuantile(99), P999: h.ValueAtQuantile(99.9), P9999: h.ValueAtQuantile(99.99),
		Max: h.Max(),
	}
}

// Export returns the compressed histogram bytes (HdrHistogram V2 log format) for reproducibility.
func (l *LatencyHist) Export() *hdrhistogram.Snapshot {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.h.Export()
}
