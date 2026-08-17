"""Paired analysis of v1 vs v2 runs.

For each cell (case, scenario, rate) it pairs repetition i of v1 with repetition i of v2 and
computes: medians, paired log-ratio effect (geometric-mean % change v2 vs v1), seeded
bootstrap 95% CI, repetition CV, gate summary. Runs that failed gates are reported but
EXCLUDED from the effect estimate (and flagged), per the DoD.

usage: python scripts/analyze.py results/ [--out analysis/]
"""
import json, glob, os, sys, math, random, statistics as st, re, csv, datetime as dt

ARM_ORDER = ["v1", "v2", "postgres", "horizon"]
SOFT_GATES = set()  # gates treated as warnings (documented), e.g. for exp A where G1/G2 cannot hold for HorizonDB


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
        m = re.search(r"[\\/](C\d)[\\/](S\d)[\\/]rate(\d+)[\\/]rep(\d+)-\d+T\d+Z-(\w+)\.json$", f)
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


def fix_g1(r):
    """Recompute G1 from env when the runner could not (PG shared_buffers parse bug in early binary)."""
    env = r.get("env") or {}
    v = env.get("variables", {})
    bp = v.get("innodb_buffer_pool_size")
    if not bp and v.get("shared_buffers"):
        sb = v["shared_buffers"]
        mult = {"kB": 1024, "MB": 1 << 20, "GB": 1 << 30, "TB": 1 << 40}
        for u, m in mult.items():
            if sb.endswith(u):
                bp = str(int(sb[:-len(u)]) * m); break
    if bp:
        ratio = env.get("total_gib", 0) / (float(bp) / 2**30)
        r["gates"]["G1_dataset_vs_bufferpool"] = {"pass": ratio >= 2.0, "value": f"{env.get('total_gib',0):.1f} GiB / {float(bp)/2**30:.1f} GiB = {ratio:.2f}x", "rule": ">= 2.0x"}
        r["gates_passed"] = all(g["pass"] for g in r["gates"].values())


def run_ok(r):
    fix_g1(r)
    hard = {k: v for k, v in r["gates"].items() if k not in SOFT_GATES}
    g = all(v["pass"] for v in hard.values())
    azok = True if r["_azmon"] is None else r["_azmon"]["gate_G6_platform"]["pass"]
    invok = True if r["_inv"] is None else r["_inv"]["violations"] == 0
    soft_failed = [k for k in SOFT_GATES if k in r["gates"] and not r["gates"][k]["pass"]]
    return g and azok and invok, {"gates": g, "azmon_G6": azok, "invariants_G9": invok, "azmon_present": r["_azmon"] is not None, "soft_failed": soft_failed}


def analyze(root, out):
    runs = load_runs(root)
    os.makedirs(out, exist_ok=True)
    rows = []
    report = {"generated_at": dt.datetime.now(dt.timezone.utc).isoformat(), "cells": []}
    md = ["# v1 vs v2 paired analysis", "", f"Generated {report['generated_at']}", ""]
    for (case, scen, rate), reps in sorted(runs.items()):
        cell = {"case": case, "scenario": scen, "rate": rate, "reps": {}, "outcomes": {}}
        usable = []
        arms_present = sorted({a for r_ in reps.values() for a in r_}, key=lambda a: ARM_ORDER.index(a) if a in ARM_ORDER else 99)
        A, B = (arms_present + [None, None])[:2]
        cell["arms"] = [A, B]
        for rep, arms in sorted(reps.items()):
            info = {}
            for arm in (A, B):
                if arm in arms:
                    ok, why = run_ok(arms[arm])
                    info[arm] = {"ok": ok, **why, "gates": {k: v["pass"] for k, v in arms[arm]["gates"].items()},
                                 "tps": arms[arm]["success_tps"], "p99_us": arms[arm]["overall"]["latency"]["p99_us"],
                                 "measure_from": arms[arm]["measure_from"], "measure_to": arms[arm]["measure_to"]}
            cell["reps"][rep] = info
            if A in arms and B in arms and info[A]["ok"] and info[B]["ok"]:
                usable.append((rep, arms[A], arms[B]))
        cell["n_pairs_total"] = sum(1 for a in reps.values() if A in a and B in a)
        cell["n_pairs_usable"] = len(usable)
        md += [f"## {case} / {scen} / {rate} arrivals/s  ({A} vs {B})", "", f"pairs: {cell['n_pairs_usable']} usable of {cell['n_pairs_total']} (hard-gate-failed pairs excluded; soft gates: {sorted(SOFT_GATES) or 'none'})", "",
               f"| Outcome | n | {A} median | {B} median | {A} CV% | {B} CV% | Δ% ({B} vs {A}, geo-mean) | 95% CI | note |", "|---|---:|---:|---:|---:|---:|---:|---|---|"]
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
                    o["note"] = f"{B} lower" + (" (better)" if better == "lower" else " (worse)")
                elif lo > 0:
                    o["note"] = f"{B} higher" + (" (better)" if better == "higher" else " (worse)")
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
                md.append(f"- rep {rep} {arm}: ok={i['ok']} tps={i['tps']:.0f} p99={i['p99_us']}us azmon={i['azmon_G6']} inv={i['invariants_G9']}" + (f" FAILED: {failed}" if failed else "") + (f" (soft-gate warnings: {i['soft_failed']})" if i.get('soft_failed') else ""))
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
    if "--soft" in sys.argv:
        SOFT_GATES.update(sys.argv[sys.argv.index("--soft") + 1].split(","))
    analyze(root, out)
