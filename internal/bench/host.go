package bench

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HostSampler records load-generator CPU% and NIC bytes/s from /proc (Linux).
// On non-Linux hosts it records nothing and reports Available=false, which the
// gate logic treats as a FAILURE of G3 (never a vacuous pass).
type HostSample struct {
	T       time.Time `json:"t"`
	CPUPct  float64   `json:"cpu_pct"`
	RxBps   float64   `json:"rx_bps"`
	TxBps   float64   `json:"tx_bps"`
	Sockets int       `json:"tcp_sockets"`
}

type HostSampler struct {
	interval time.Duration
	mu       sync.Mutex
	samples  []HostSample
	stop     chan struct{}
	done     chan struct{}
	Available bool
}

func NewHostSampler(interval time.Duration) *HostSampler {
	_, err := os.Stat("/proc/stat")
	return &HostSampler{interval: interval, stop: make(chan struct{}), done: make(chan struct{}), Available: err == nil}
}

func readCPU() (idle, total uint64, ok bool) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 5 || fields[0] != "cpu" {
			continue
		}
		var vals []uint64
		for _, s := range fields[1:] {
			v, _ := strconv.ParseUint(s, 10, 64)
			vals = append(vals, v)
		}
		for i, v := range vals {
			total += v
			if i == 3 || i == 4 {
				idle += v
			}
		}
		return idle, total, true
	}
	return 0, 0, false
}

func readNet() (rx, tx uint64) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		i := strings.Index(line, ":")
		if i < 0 {
			continue
		}
		name := strings.TrimSpace(line[:i])
		if name == "lo" {
			continue
		}
		fields := strings.Fields(line[i+1:])
		if len(fields) < 9 {
			continue
		}
		r, _ := strconv.ParseUint(fields[0], 10, 64)
		t, _ := strconv.ParseUint(fields[8], 10, 64)
		rx += r
		tx += t
	}
	return
}

func countTCP() int {
	n := 0
	for _, p := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		sc.Scan() // header
		for sc.Scan() {
			n++
		}
		f.Close()
	}
	return n
}

func (h *HostSampler) Start() {
	if !h.Available {
		close(h.done)
		return
	}
	go func() {
		defer close(h.done)
		pi, pt, _ := readCPU()
		prx, ptx := readNet()
		last := time.Now()
		t := time.NewTicker(h.interval)
		defer t.Stop()
		for {
			select {
			case <-h.stop:
				return
			case now := <-t.C:
				i, tot, ok := readCPU()
				rx, tx := readNet()
				if !ok {
					continue
				}
				dt := now.Sub(last).Seconds()
				cpu := 0.0
				if tot > pt {
					cpu = 100 * (1 - float64(i-pi)/float64(tot-pt))
				}
				s := HostSample{T: now.UTC(), CPUPct: cpu, RxBps: float64(rx-prx) / dt, TxBps: float64(tx-ptx) / dt, Sockets: countTCP()}
				h.mu.Lock()
				h.samples = append(h.samples, s)
				h.mu.Unlock()
				pi, pt, prx, ptx, last = i, tot, rx, tx, now
			}
		}
	}()
}

func (h *HostSampler) Stop() []HostSample {
	if h.Available {
		close(h.stop)
	}
	<-h.done
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.samples
}

type HostSummary struct {
	Available bool    `json:"available"`
	CPUMax    float64 `json:"cpu_pct_max"`
	CPUAvg    float64 `json:"cpu_pct_avg"`
	RxMaxBps  float64 `json:"rx_bps_max"`
	TxMaxBps  float64 `json:"tx_bps_max"`
	Samples   int     `json:"samples"`
	NICLimitBps float64 `json:"nic_limit_bps_assumed"`
	NetMaxPct float64 `json:"net_pct_max"`
}

func SummarizeHost(samples []HostSample, from, to time.Time, nicLimitBps float64) HostSummary {
	s := HostSummary{Available: len(samples) > 0, NICLimitBps: nicLimitBps}
	sum := 0.0
	for _, x := range samples {
		if x.T.Before(from) || x.T.After(to) {
			continue
		}
		s.Samples++
		sum += x.CPUPct
		if x.CPUPct > s.CPUMax {
			s.CPUMax = x.CPUPct
		}
		if x.RxBps > s.RxMaxBps {
			s.RxMaxBps = x.RxBps
		}
		if x.TxBps > s.TxMaxBps {
			s.TxMaxBps = x.TxBps
		}
	}
	if s.Samples > 0 {
		s.CPUAvg = sum / float64(s.Samples)
	} else {
		s.Available = false
	}
	if nicLimitBps > 0 {
		m := s.RxMaxBps
		if s.TxMaxBps > m {
			m = s.TxMaxBps
		}
		s.NetMaxPct = 100 * m / nicLimitBps
	}
	return s
}
