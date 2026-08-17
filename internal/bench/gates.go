package bench

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
)

type GateResult struct {
	Pass   bool   `json:"pass"`
	Value  string `json:"value"`
	Rule   string `json:"rule"`
	Detail string `json:"detail,omitempty"`
}

// Gate thresholds (from benchmark-prompt-v2.md §3). Exposed so orchestrators can tighten.
var (
	G1DatasetToBufferPool = 2.0
	G2MaxHitRatio         = 0.99
	G3MaxCPUPct           = 60.0
	G3MaxNetPct           = 50.0
	G4MaxQueueP99Us       = int64(5000) // queue pickup delay p99 (workers == connections, no pool wait)
	G5MaxErrorRate        = 0.01
	G5MaxTimeoutRate      = 0.005
	G5MaxDropRate         = 0.001
	G6MinStatusSamples    = 3
	G10MinDurationRatio   = 0.95
)

func EvaluateGates(r *RunResult) (map[string]GateResult, bool) {
	g := map[string]GateResult{}
	all := true
	add := func(k string, pass bool, val, rule, detail string) {
		g[k] = GateResult{Pass: pass, Value: val, Rule: rule, Detail: detail}
		if !pass {
			all = false
		}
	}
	// G1 dataset vs buffer pool
	if r.Env != nil {
		bp, _ := strconv.ParseFloat(r.Env.Variables["innodb_buffer_pool_size"], 64)
		bpGiB := bp / 1073741824
		ratio := 0.0
		if bpGiB > 0 {
			ratio = r.Env.TotalGiB / bpGiB
		}
		add("G1_dataset_vs_bufferpool", ratio >= G1DatasetToBufferPool, fmt.Sprintf("%.1f GiB / %.1f GiB = %.2fx", r.Env.TotalGiB, bpGiB, ratio), fmt.Sprintf(">= %.1fx", G1DatasetToBufferPool), "")
	} else {
		add("G1_dataset_vs_bufferpool", false, "no env snapshot", ">= 2x", "")
	}
	// G2 storage actually used
	if r.StatusDelta != nil {
		d := r.StatusDelta
		add("G2_storage_used", d.ReadIOPS > 0 && d.BufferPoolHitRatio < G2MaxHitRatio,
			fmt.Sprintf("read_iops=%.1f write_iops=%.1f hit=%.4f", d.ReadIOPS, d.WriteIOPS, d.BufferPoolHitRatio),
			fmt.Sprintf("read_iops>0 && hit<%.2f", G2MaxHitRatio), "")
	} else {
		add("G2_storage_used", false, "no server status delta", "read_iops>0", "")
	}
	// G3 load generator headroom
	h := r.Host
	add("G3_client_headroom", h.Available && h.CPUMax < G3MaxCPUPct && (h.NICLimitBps == 0 || h.NetMaxPct < G3MaxNetPct),
		fmt.Sprintf("available=%v cpu_max=%.1f%% net_max=%.1f%%", h.Available, h.CPUMax, h.NetMaxPct),
		fmt.Sprintf("measured && cpu<%.0f%% && net<%.0f%%", G3MaxCPUPct, G3MaxNetPct), "")
	// G4 no pool wait: dedicated connections; queue pickup delay p99 must be small
	add("G4_queue_delay", r.QueueDelay.P99 < G4MaxQueueP99Us,
		fmt.Sprintf("queue_p99=%dus max_inflight=%d/%d", r.QueueDelay.P99, r.MaxInFlight, r.Config.Workers),
		fmt.Sprintf("p99<%dus", G4MaxQueueP99Us), "")
	// G5 errors
	add("G5_errors", r.ErrorRate < G5MaxErrorRate && r.TimeoutRate < G5MaxTimeoutRate && r.DropRate < G5MaxDropRate,
		fmt.Sprintf("err=%.4f%% timeout=%.4f%% drop=%.4f%%", 100*r.ErrorRate, 100*r.TimeoutRate, 100*r.DropRate),
		fmt.Sprintf("err<%.1f%% timeout<%.1f%% drop<%.1f%%", 100*G5MaxErrorRate, 100*G5MaxTimeoutRate, 100*G5MaxDropRate), "")
	// G6 server metrics present & aligned (in-run status sampler; Azure Monitor slice is checked by orchestrator)
	n := 0
	if r.StatusDelta != nil {
		n = r.StatusDelta.Samples
	}
	add("G6_server_metrics", n >= G6MinStatusSamples && r.StatusErrors == 0,
		fmt.Sprintf("status_samples=%d errors=%d", n, r.StatusErrors), fmt.Sprintf(">=%d samples, 0 errors, UTC", G6MinStatusSamples), "")
	// G7 steady state
	add("G7_steady_state", r.SteadyState.Reached, fmt.Sprintf("cv=%.2f%% warmup=%ds", r.SteadyState.CVPct, r.SteadyState.WarmupSecUsed), fmt.Sprintf("cv<%.1f%%", r.Config.SteadyCVPct), "")
	// G10 duration
	dur := r.MeasureTo.Sub(r.MeasureFrom).Seconds()
	ratio := dur / float64(r.Config.MeasureSec)
	add("G10_duration", ratio >= G10MinDurationRatio, fmt.Sprintf("%.1fs / %ds", dur, r.Config.MeasureSec), ">=95%", "")
	return g, all
}

func encodeHist(h *LatencyHist) string {
	snap := h.Export()
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, snap.LowestTrackableValue)
	binary.Write(&buf, binary.LittleEndian, snap.HighestTrackableValue)
	binary.Write(&buf, binary.LittleEndian, snap.SignificantFigures)
	binary.Write(&buf, binary.LittleEndian, int64(len(snap.Counts)))
	for _, c := range snap.Counts {
		binary.Write(&buf, binary.LittleEndian, c)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
