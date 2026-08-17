"""Slice Azure Monitor platform metrics for each run's UTC measurement window and attach them
next to the run JSON as <run>.azmon.json. Also verifies (gate G6-platform) that the slice is
non-empty and aligned with the run window.

usage: python scripts/azmon_slice.py results/<case>/<scenario>/rate*/rep*-<arm>.json ...
"""
import json, subprocess, sys, glob, os, datetime as dt

SUB = "7784c8b4-64ba-4b09-bee8-ee6f8f9a7309"
RG = "mysql-storage-benchmark"
SERVERS = {"v1": "mysqlbm-euson-v1", "v2": "mysqlbm-euson-v2"}
METRICS = ["cpu_percent", "memory_percent", "io_consumption_percent", "storage_io_count", "active_connections",
           "Queries", "Innodb_buffer_pool_reads", "Innodb_buffer_pool_read_requests", "Innodb_data_writes",
           "network_bytes_ingress", "network_bytes_egress", "aborted_connections", "total_connections", "storage_percent"]


def az(args):
    p = subprocess.run(["az"] + args, capture_output=True, text=True, shell=True)
    if p.returncode != 0:
        raise RuntimeError(p.stderr[:500])
    return json.loads(p.stdout) if p.stdout.strip() else None


def slice_run(path):
    run = json.load(open(path, encoding="utf-8"))
    arm = run["config"]["label"]
    server = SERVERS.get(arm)
    if not server:
        print("skip (unknown arm)", path); return
    t0 = dt.datetime.fromisoformat(run["measure_from"].replace("Z", "+00:00")) - dt.timedelta(minutes=1)
    t1 = dt.datetime.fromisoformat(run["measure_to"].replace("Z", "+00:00")) + dt.timedelta(minutes=1)
    rid = f"/subscriptions/{SUB}/resourceGroups/{RG}/providers/Microsoft.DBforMySQL/flexibleServers/{server}"
    fmt = "%Y-%m-%dT%H:%M:%SZ"
    out = {"server": server, "window_utc": [t0.strftime(fmt), t1.strftime(fmt)], "metrics": {}, "nonempty": {}, "captured_at": dt.datetime.now(dt.timezone.utc).isoformat()}
    for m in METRICS:
        try:
            r = az(["monitor", "metrics", "list", "--resource", rid, "--metric", m, "--start-time", t0.strftime(fmt), "--end-time", t1.strftime(fmt),
                    "--interval", "PT1M", "--aggregation", "Average", "Maximum", "Total", "-o", "json"])
        except Exception as e:
            out["metrics"][m] = {"error": str(e)}; out["nonempty"][m] = False; continue
        series = []
        for v in r.get("value", []):
            for ts in v.get("timeseries", []):
                for d in ts.get("data", []):
                    series.append({"t": d.get("timeStamp"), "avg": d.get("average"), "max": d.get("maximum"), "total": d.get("total")})
        out["metrics"][m] = series
        out["nonempty"][m] = any(x["avg"] is not None or x["total"] is not None for x in series)
    core = ["cpu_percent", "io_consumption_percent", "active_connections"]
    out["gate_G6_platform"] = {"pass": all(out["nonempty"].get(m) for m in core), "core_metrics": {m: out["nonempty"].get(m) for m in core}}
    dst = path[:-5] + ".azmon.json"
    json.dump(out, open(dst, "w", encoding="utf-8"), indent=1)
    print(f"{os.path.basename(path)}: G6-platform={out['gate_G6_platform']['pass']} -> {os.path.basename(dst)}")


if __name__ == "__main__":
    files = []
    for a in sys.argv[1:]:
        files += glob.glob(a)
    for f in files:
        if f.endswith(".azmon.json") or f.endswith(".invariants.json"):
            continue
        slice_run(f)
