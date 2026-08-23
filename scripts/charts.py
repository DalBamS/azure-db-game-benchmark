"""Charts from run JSONs: per-second throughput/p99 overlaid with server read IOPS (same UTC axis),
latency percentile bars per cell, and knee staircase. Output PNGs to analysis/charts/.

usage: python scripts/charts.py results/ analysis/charts
"""
import json, glob, os, sys, re, datetime as dt
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.dates as mdates

C = {"v1": "#1f77b4", "v2": "#d62728", "postgres": "#1f77b4", "horizon": "#d62728"}


def ts(s):
    return dt.datetime.fromisoformat(s.replace("Z", "+00:00"))


def series_status(run, key):
    xs, ys = [], []
    prev = None
    for s in run["server_status"]:
        t = ts(s["t"]); v = s["v"].get(key)
        if prev is not None and v is not None:
            dtv = (t - prev[0]).total_seconds()
            if dtv > 0:
                xs.append(t); ys.append((v - prev[1]) / dtv)
        if v is not None:
            prev = (t, v)
    return xs, ys


def timeline_chart(pair, out):
    """pair = {arm: run}"""
    fig, axes = plt.subplots(3, 1, figsize=(11, 8.5), sharex=True)
    for arm, run in pair.items():
        secs = run["per_second"]
        t = [ts(s["t"]) for s in secs]
        axes[0].plot(t, [s["success"] for s in secs], color=C.get(arm, None), lw=0.9, label=f"{arm} success/s")
        axes[1].plot(t, [s["p99_us"] / 1000 for s in secs], color=C.get(arm, None), lw=0.9, label=f"{arm} p99 (ms, 1s window)")
        xs, ys = series_status(run, "Innodb_data_reads")
        axes[2].plot(xs, ys, color=C.get(arm, None), lw=1.2, label=f"{arm} server read IOPS")
        xs, ys = series_status(run, "Innodb_data_writes")
        axes[2].plot(xs, ys, color=C.get(arm, None), lw=1.0, ls="--", label=f"{arm} server write IOPS")
        for ax in axes:
            ax.axvspan(ts(run["measure_from"]), ts(run["measure_to"]), color=C.get(arm, "gray"), alpha=0.04)
    axes[0].set_ylabel("TPS"); axes[1].set_ylabel("p99 ms"); axes[2].set_ylabel("IOPS")
    axes[1].set_yscale("log")
    for ax in axes:
        ax.grid(alpha=0.3); ax.legend(loc="upper right", fontsize=8)
    axes[2].xaxis.set_major_formatter(mdates.DateFormatter("%H:%M"))
    r0 = next(iter(pair.values()))
    fig.suptitle(f"{r0['config']['scenario']} @ {r0['config']['target_rate']:.0f}/s — client TPS / p99 vs server IOPS (UTC, shaded = measurement window)")
    fig.tight_layout()
    fig.savefig(out, dpi=130); plt.close(fig)


def percentile_chart(cells, out):
    """cells: {label: {arm: [runs]}}"""
    labels = list(cells.keys())
    pcts = ["p50_us", "p95_us", "p99_us", "p999_us"]
    ncol = min(3, len(labels))
    nrow = (len(labels) + ncol - 1) // ncol
    fig, axes = plt.subplots(nrow, ncol, figsize=(5 * ncol, 4.0 * nrow), squeeze=False)
    for i, lab in enumerate(labels):
        ax = axes[i // ncol][i % ncol]
        arms = sorted(cells[lab].keys())
        w = 0.38
        for j, arm in enumerate(arms):
            runs = cells[lab][arm]
            med = [sorted(r["overall"]["latency"][p] for r in runs)[len(runs) // 2] / 1000 for p in pcts]
            lo = [min(r["overall"]["latency"][p] for r in runs) / 1000 for p in pcts]
            hi = [max(r["overall"]["latency"][p] for r in runs) / 1000 for p in pcts]
            x = [k + (j - 0.5) * w for k in range(len(pcts))]
            ax.bar(x, med, width=w, color=C.get(arm), label=f"{arm} (median of {len(runs)} reps)")
            ax.errorbar(x, med, yerr=[[m - l for m, l in zip(med, lo)], [h - m for m, h in zip(med, hi)]], fmt="none", ecolor="black", capsize=3, lw=0.8)
        ax.set_xticks(range(len(pcts))); ax.set_xticklabels(["p50", "p95", "p99", "p99.9"])
        ax.set_yscale("log"); ax.set_ylabel("latency ms (log)"); ax.set_title(lab); ax.grid(alpha=0.3, axis="y"); ax.legend(fontsize=8)
    for j in range(len(labels), nrow * ncol):
        axes[j // ncol][j % ncol].axis("off")
    fig.tight_layout(); fig.savefig(out, dpi=130); plt.close(fig)


def knee_chart(files, out):
    fig, ax1 = plt.subplots(figsize=(8, 4.5))
    ax2 = ax1.twinx()
    for f in files:
        d = json.load(open(f, encoding="utf-8"))
        arm = d["label"]
        xs = [s["workers"] for s in d["steps"]]
        ax1.plot(xs, [s["success_tps"] for s in d["steps"]], "o-", color=C.get(arm), label=f"{arm} TPS")
        ax2.plot(xs, [s["p99_us"] / 1000 for s in d["steps"]], "s--", color=C.get(arm), alpha=0.6, label=f"{arm} p99 ms")
        ax1.axvline(d["knee_workers"], color=C.get(arm), alpha=0.3, ls=":")
    ax1.set_xscale("log", base=2); ax1.set_xlabel("closed-loop concurrency"); ax1.set_ylabel("TPS"); ax2.set_ylabel("p99 ms"); ax2.set_yscale("log")
    ax1.grid(alpha=0.3); ax1.legend(loc="upper left", fontsize=8); ax2.legend(loc="lower right", fontsize=8)
    ax1.set_title("Knee search (closed-loop staircase, " + os.path.basename(os.path.dirname(files[0])) + ")")
    fig.tight_layout(); fig.savefig(out, dpi=130); plt.close(fig)


def main(root, out):
    os.makedirs(out, exist_ok=True)
    cells = {}
    for f in glob.glob(os.path.join(root, "**", "rep*-*.json"), recursive=True):
        if f.endswith(".azmon.json") or f.endswith(".invariants.json"):
            continue
        m = re.search(r"[\\/](C\d)[\\/](S\d)[\\/]rate(\d+)[\\/]rep(\d+)-(\d+T\d+Z)-(\w+)\.json$", f)
        if not m:
            continue
        case, scen, rate, rep, stamp, arm = m.groups()
        r = json.load(open(f, encoding="utf-8"))
        cells.setdefault(f"{case}/{scen}/{rate}", {}).setdefault(arm, []).append(r)
        cells[f"{case}/{scen}/{rate}"].setdefault("_pairs", {}).setdefault(rep, {})[arm] = r
    for lab, arms in cells.items():
        pairs = arms.pop("_pairs", {})
        for rep, pair in pairs.items():
            if len(pair) >= 2:
                timeline_chart(pair, os.path.join(out, f"timeline_{lab.replace('/', '_')}_rep{rep}.png"))
    if cells:
        percentile_chart({k: v for k, v in cells.items()}, os.path.join(out, "latency_percentiles.png"))
    for d in {os.path.dirname(f) for f in glob.glob(os.path.join(root, "**", "knee-*.json"), recursive=True)}:
        fs = sorted(glob.glob(os.path.join(d, "knee-*.json")))
        knee_chart(fs, os.path.join(out, "knee_" + "_".join(d.replace("\\", "/").split("/")[-2:]) + ".png"))
    print("charts ->", out, os.listdir(out))


if __name__ == "__main__":
    main(sys.argv[1] if len(sys.argv) > 1 else "results", sys.argv[2] if len(sys.argv) > 2 else "analysis/charts")
