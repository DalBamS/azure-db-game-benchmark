"""Paired analysis of v1 vs v2 runs.

For each cell (case, scenario, rate) it pairs repetition i of v1 with repetition i of v2 and
computes: medians, paired log-ratio effect (geometric-mean % change v2 vs v1), seeded
bootstrap 95% CI, repetition CV, gate summary. Runs that failed gates are reported but
EXCLUDED from the effect estimate (and flagged), per the DoD.

usage: python scripts/analyze.py results/ [--out analysis/]
"""
import json, glob, os, sys, math, random, statistics as st, re, csv, datetime as dt

OUTCOMES = [
    ("success_tps", "Throughput (success TPS)", "higher", lambda r: r["success_tps"]),
    ("p50_us", "Latency p50 (us)", "lower", lambda r: r["overall"]["latency"]["p50_us"]),
    ("p95_us", "Latency p95 (us)", "lower", lambda r: r["overall"]["latency"]["p95_us"]),
    ("p99_us", "Latency p99 (us)", "lower", lambda r: r["overall"]["latency"]["p99_us"]),
    ("p999_us", "Latency p99.9 (us)", "lower", lambda r: r["overall"]["latency"]["p999_us"]),
    ("error_rate", "Error rate", "lower", lambda r: r["error_rate"]),
    ("read_iops", "Server read IOPS", None, lambda r: (r.get("server_status_delta") or {}).get("read_iops")),
    ("write_iops", "Server write IOPS", None, lambda r: (r.get("server_status_delta") or {}).get("write_iops")),
    ("bp_hit", "Buffer pool hit ratio", None, lambda r: (r.get("server_status_delta") or {}).get("buffer_pool_hit_ratio")),
    ("service_p99_us", "Service time p99 (us, excl. queue)", "lower", lambda r: r["service_time"]["p99_us"]),
]


def bootstrap_ci(logs, n=10000, seed=1234):
    if len(logs) < 2:
        return None
    rnd = random.Random(seed)
    means = []
    k = len(logs)
    for _ in range(n):
        s = sum(logs[rnd.randrange(k)] for _ in range(k)) / k
        means.append(s)
    means.sort()
    return means[int(0.025 * n)], means[int(0.975 * n) - 1]


def cv(xs):
    if len(xs) < 2 or st.mean(xs) == 0:
        return None
    return 100 * st.stdev(xs) / st.mean(xs)


def load_runs(root):
    runs = {}
    for f in glob.glob(os.path.join(root, "**", "rep*-*.json"), recursive=True):
        if f.endswith(".azmon.json") or f.endswith(".invariants.json"):
            continue
        m = re.search(r"[\\/](C\d)[\\/](S\d)[\\/]rate(\d+)[\\/]rep(\d+)-\d+T\d+Z-(v\d)\.json$", f)
        if not m:
            continue
        case, scen, rate, rep, arm = m.groups()
        r = json.load(open(f, encoding="utf-8"))
        az = f[:-5] + ".azmon.json"
        r["_azmon"] = json.load(open(az, encoding="utf-8")) if os.path.exists(az) else None
        inv = f[:-5] + ".invariants.json"
        r["_inv"] = json.load(open(inv, encoding="utf-8")) if os.path.exists(inv) else None
        r["_file"] = f
        runs.setdefault((case, scen, int(rate)), {}).setdefault(int(rep), {})[arm] = r
    return runs


def run_ok(r):
    g = r.get("gates_passed", False)
    azok = True if r["_azmon"] is None else r["_azmon"]["gate_G6_platform"]["pass"]
    invok = True if r["_inv"] is None else r["_inv"]["violations"] == 0
    return g and azok and invok, {"gates": g, "azmon_G6": azok, "invariants_G9": invok, "azmon_present": r["_azmon"] is not None}


def analyze(root, out):
    runs = load_runs(root)
    os.makedirs(out, exist_ok=True)
    rows = []
    report = {"generated_at": dt.datetime.now(dt.timezone.utc).isoformat(), "cells": []}
    md = ["# v1 vs v2 paired analysis", "", f"Generated {report['generated_at']}", ""]
    for (case, scen, rate), reps in sorted(runs.items()):
        cell = {"case": case, "scenario": scen, "rate": rate, "reps": {}, "outcomes": {}}
        usable = []
        for rep, arms in sorted(reps.items()):
            info = {}
            for arm in ("v1", "v2"):
                if arm in arms:
                    ok, why = run_ok(arms[arm])
                    info[arm] = {"ok": ok, **why, "gates": {k: v["pass"] for k, v in arms[arm]["gates"].items()},
                                 "tps": arms[arm]["success_tps"], "p99_us": arms[arm]["overall"]["latency"]["p99_us"],
                                 "measure_from": arms[arm]["measure_from"], "measure_to": arms[arm]["measure_to"]}
            cell["reps"][rep] = info
            if "v1" in arms and "v2" in arms and info["v1"]["ok"] and info["v2"]["ok"]:
                usable.append((rep, arms["v1"], arms["v2"]))
        cell["n_pairs_total"] = sum(1 for a in reps.values() if "v1" in a and "v2" in a)
        cell["n_pairs_usable"] = len(usable)
        md += [f"## {case} / {scen} / {rate} arrivals/s", "", f"pairs: {cell['n_pairs_usable']} usable of {cell['n_pairs_total']} (gate-failed pairs excluded)", "",
               "| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |", "|---|---:|---:|---:|---:|---:|---:|---|---|"]
        for key, label, better, fn in OUTCOMES:
            v1s = [fn(a) for _, a, _ in usable]
            v2s = [fn(b) for _, _, b in usable]
            v1s = [x for x in v1s if x is not None]; v2s = [x for x in v2s if x is not None]
            o = {"label": label, "n": min(len(v1s), len(v2s)), "v1_median": st.median(v1s) if v1s else None, "v2_median": st.median(v2s) if v2s else None,
                 "v1_cv_pct": cv(v1s), "v2_cv_pct": cv(v2s), "pct_change": None, "ci95": None, "note": ""}
            pairs = [(a, b) for a, b in zip(v1s, v2s) if a and b and a > 0 and b > 0]
            if better and len(pairs) >= 2:
                logs = [math.log(b / a) for a, b in pairs]
                mean = sum(logs) / len(logs)
                o["pct_change"] = 100 * (math.exp(mean) - 1)
                ci = bootstrap_ci(logs)
                o["ci95"] = [100 * (math.exp(ci[0]) - 1), 100 * (math.exp(ci[1]) - 1)]
                lo, hi = o["ci95"]
                if hi < 0:
                    o["note"] = "v2 lower" + (" (better)" if better == "lower" else " (worse)")
                elif lo > 0:
                    o["note"] = "v2 higher" + (" (better)" if better == "higher" else " (worse)")
                else:
                    o["note"] = "CI includes 0"
                if len(pairs) < 5:
                    o["note"] += f"; n={len(pairs)}<5 (below protocol baseline)"
            elif better:
                o["note"] = "insufficient pairs for effect/CI"
            cell["outcomes"][key] = o
            f = lambda x: "-" if x is None else (f"{x:.4f}" if abs(x) < 1 else f"{x:,.1f}")
            md.append(f"| {label} | {o['n']} | {f(o['v1_median'])} | {f(o['v2_median'])} | {f(o['v1_cv_pct'])} | {f(o['v2_cv_pct'])} | "
                      f"{'-' if o['pct_change'] is None else f'{o['pct_change']:+.2f}%'} | "
                      f"{'-' if o['ci95'] is None else f'[{o['ci95'][0]:+.2f}%, {o['ci95'][1]:+.2f}%]'} | {o['note']} |")
            rows.append({"case": case, "scenario": scen, "rate": rate, "outcome": key, **{k: v for k, v in o.items() if k != "label"}})
        md.append("")
        md.append("Repetition gate summary:")
        for rep, info in sorted(cell["reps"].items()):
            for arm, i in info.items():
                failed = [k for k, v in i["gates"].items() if not v]
                md.append(f"- rep {rep} {arm}: ok={i['ok']} tps={i['tps']:.0f} p99={i['p99_us']}us azmon={i['azmon_G6']} inv={i['invariants_G9']}" + (f" FAILED: {failed}" if failed else ""))
        md.append("")
        report["cells"].append(cell)
    json.dump(report, open(os.path.join(out, "analysis.json"), "w", encoding="utf-8"), indent=1)
    open(os.path.join(out, "analysis.md"), "w", encoding="utf-8").write("\n".join(md))
    if rows:
        with open(os.path.join(out, "analysis.csv"), "w", newline="", encoding="utf-8") as fh:
            w = csv.DictWriter(fh, fieldnames=list(rows[0].keys()))
            w.writeheader(); w.writerows(rows)
    print("\n".join(md))


if __name__ == "__main__":
    root = sys.argv[1] if len(sys.argv) > 1 else "results"
    out = sys.argv[sys.argv.index("--out") + 1] if "--out" in sys.argv else "analysis"
    analyze(root, out)
