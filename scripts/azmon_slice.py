"""Slice Azure Monitor platform metrics for each run's UTC measurement window and attach them
next to the run JSON as <run>.azmon.json. Also verifies (gate G6-platform) that the slice is
non-empty and aligned with the run window.

usage: python scripts/azmon_slice.py results/<case>/<scenario>/rate*/rep*-<arm>.json ...
"""
import json, subprocess, sys, glob, os, datetime as dt

SUB = "7784c8b4-64ba-4b09-bee8-ee6f8f9a7309"
RG = "mysql-storage-benchmark"
SERVERS = {
    "v1": (RG, "Microsoft.DBforMySQL/flexibleServers", "mysqlbm-euson-v1"),
    "v2": (RG, "Microsoft.DBforMySQL/flexibleServers", "mysqlbm-euson-v2"),
    "postgres": ("rg-expa-pg-hz", "Microsoft.DBforPostgreSQL/flexibleServers", "expa-pg"),
    "horizon": ("rg-expa-pg-hz", "Microsoft.HorizonDB/clusters", "expa-hz"),
}
METRICS_BY_TYPE = {
    "Microsoft.DBforMySQL/flexibleServers": ["cpu_percent", "memory_percent", "io_consumption_percent", "storage_io_count", "active_connections",
        "Queries", "Innodb_buffer_pool_reads", "Innodb_buffer_pool_read_requests", "Innodb_data_writes", "network_bytes_ingress", "network_bytes_egress", "aborted_connections", "total_connections", "storage_percent"],
    "Microsoft.DBforPostgreSQL/flexibleServers": ["cpu_percent", "memory_percent", "iops", "read_iops", "write_iops", "disk_iops_consumed_percentage", "disk_bandwidth_consumed_percentage", "disk_queue_depth", "active_connections",
        "network_bytes_ingress", "network_bytes_egress", "storage_percent", "read_throughput", "write_throughput", "blks_read", "blks_hit", "tps", "physical_replication_delay_in_seconds"],
    "Microsoft.HorizonDB/clusters": ["CpuPercent", "MemoryPercent", "ActiveConnections", "NetworkBytesIngress", "NetworkBytesEgress", "TransactionsPerSecond", "CommitLatency", "WriteLatency", "WALBytesWritten", "WALWritesPerSecond", "BufferPoolCacheHitRatio", "StorageUsed", "Deadlocks", "TuplesFetched", "TuplesUpdated"],
}
CORE_BY_TYPE = {"Microsoft.DBforMySQL/flexibleServers": ["cpu_percent", "io_consumption_percent", "active_connections"],
                "Microsoft.DBforPostgreSQL/flexibleServers": ["cpu_percent", "iops", "active_connections"],
                "Microsoft.HorizonDB/clusters": ["CpuPercent", "ActiveConnections"]}


def az(args):
    p = subprocess.run(["az"] + args, capture_output=True, text=True, shell=True)
    if p.returncode != 0:
        raise RuntimeError(p.stderr[:500])
    return json.loads(p.stdout) if p.stdout.strip() else None


def slice_run(path):
    run = json.load(open(path, encoding="utf-8"))
    arm = run["config"]["label"]
    if arm not in SERVERS:
        print("skip (unknown arm)", path); return
    rg, rtype, server = SERVERS[arm]
    METRICS = METRICS_BY_TYPE[rtype]
    t0 = dt.datetime.fromisoformat(run["measure_from"].replace("Z", "+00:00")) - dt.timedelta(minutes=1)
    t1 = dt.datetime.fromisoformat(run["measure_to"].replace("Z", "+00:00")) + dt.timedelta(minutes=1)
    rid = f"/subscriptions/{SUB}/resourceGroups/{rg}/providers/{rtype}/{server}"
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
    core = CORE_BY_TYPE[rtype]
    out["gate_G6_platform"] = {"pass": all(out["nonempty"].get(m) for m in core), "core_metrics": {m: out["nonempty"].get(m) for m in core}}
    dst = path[:-5] + ".azmon.json"
    json.dump(out, open(dst, "w", encoding="utf-8"), indent=1)
    print(f"{os.path.basename(path)}: G6-platform={out['gate_G6_platform']['pass']} -> {os.path.basename(dst)}")


if __name__ == "__main__":
    # --map v1=serverA,v2=serverB overrides the arm->server mapping (same RG/type as the base arm name)
    if "--map" in sys.argv:
        i = sys.argv.index("--map")
        for pair in sys.argv[i + 1].split(","):
            arm, srv = pair.split("=")
            base = SERVERS.get(arm) or SERVERS["v1"]
            SERVERS[arm] = (base[0], base[1], srv)
        del sys.argv[i:i + 2]
    files = []
    for a in sys.argv[1:]:
        files += glob.glob(a)
    for f in files:
        if f.endswith(".azmon.json") or f.endswith(".invariants.json"):
            continue
        slice_run(f)
