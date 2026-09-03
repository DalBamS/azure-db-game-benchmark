# -*- coding: utf-8 -*-
"""expA(PG vs HorizonDB) 측정 JSON을 InfluxDB line protocol로 변환한다.

- bench_client: 초당 성공 TPS / 오류 / 1초창 p99 (tags: exp,arm,cell,rep)
- bench_server: 서버 내부 지표를 구간 rate로 환산 (read/write IOPS, commits/s)

usage: python scripts/load_history_to_influx.py results/expA > /tmp/expa.lp
"""
import json, glob, os, sys, re, datetime as dt

root = sys.argv[1] if len(sys.argv) > 1 else "results/expA"
out = []

def ts_ns(s):
    return int(dt.datetime.fromisoformat(s.replace("Z", "+00:00")).timestamp() * 1e9)

for f in sorted(glob.glob(os.path.join(root, "**", "rep*-*.json"), recursive=True)):
    if f.endswith(".azmon.json") or f.endswith(".invariants.json"):
        continue
    m = re.search(r"[\\/](C\d)[\\/](S\d)[\\/]rate(\d+)[\\/]rep(\d+)-\d+T\d+Z-(\w+)\.json$", f)
    if not m:
        continue
    case, scen, rate, rep, arm = m.groups()
    cell = f"{case}-{scen}-{rate}"
    d = json.load(open(f, encoding="utf-8"))
    tags = f"exp=expA,arm={arm},cell={cell},rep={rep}"
    for s in d.get("per_second", []):
        out.append(f"bench_client,{tags} success={s['success']}i,errors={s['errors']}i,p99_ms={s['p99_us']/1000.0},queue_len={s['queue_len']}i {ts_ns(s['t'])}")
    prev = None
    for s in d.get("server_status", []):
        t = ts_ns(s["t"]); v = s["v"]
        if prev is not None:
            dts = (t - prev[0]) / 1e9
            if dts > 0:
                pv = prev[1]
                def rate_of(k):
                    return max(0.0, (v.get(k, 0) - pv.get(k, 0)) / dts)
                out.append(f"bench_server,{tags} read_iops={rate_of('Innodb_data_reads')},write_iops={rate_of('Innodb_data_writes')},commits_ps={rate_of('Com_commit')} {t}")
        prev = (t, v)

sys.stdout.write("\n".join(out))
sys.stderr.write(f"lines={len(out)}\n")
