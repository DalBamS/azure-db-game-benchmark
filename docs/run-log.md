# Run log (UTC)

## 2026-08-17

- 15:59 Reused servers `mysqlbm-euson-v1` (E8ds_v5, Premium_LRS) / `-v2` (E8ds_v6, PremiumV2_LRS). Set `innodb_buffer_pool_size`=8 GiB on both (pre-registered control), storage 64→128 GiB, v1 IOPS 492→5000 (same-IOPS cell).
- 16:07 v1 HA enable failed: `AutoGrowDisabledNotSupportedForHA` → enabled autogrow on v1, retried HA (SameZone). v2 HA SameZone enabled OK (autoGrow reported Disabled but accepted).
- 16:03 Bench VM `mysqlbm-euson-lg-vm` resized D4ds_v5→D16ds_v5, public IP + NSG (SSH/3000 from operator IP only). Key Vault (private, RBAC) not readable by operator → MySQL password must be placed on the VM by the user (`scripts/vm-fetch-mysql-secret.sh` via run-command).
- 16:15 Monitoring VM `bench-mon-vm` (D4s_v5) created; InfluxDB2 + Grafana + Telegraf installed via `infra/monitoring/cloud-init.yaml` (executed as setup.sh). MySQL inputs collected from bench VM telegraf (credentials live there), pushed to monitoring InfluxDB.
- Time budget from user: ~7h total; scope reduced to C1 (8 vCore/SameZone HA) and C5 (16 vCore/SameZone HA), 3 reps × 10 min, arms run concurrently.
- 16:35 Exp B dataset load both arms (2.5M accounts × 20 slots, 1 KiB profile / 320 B attrs random): v1 24.4 GiB in 9.5 min, v2 23.7 GiB in ~17 min (v2 slower to bulk-load; doublewrite ON on v2).
- 16:45 Smoke on v1 (500/s, 60 s): all gates pass; read IOPS 345, BP hit 95.2%.
- 16:53 Knee C1/S1 (closed-loop 32..512, 90 s/step, arms concurrent): v1 knee@64 (~7.9k TPS), v2 knee@256 (~9.6k TPS); server read IOPS plateau ~2.7k both. Common measurement rate = 5000/s (≈65% of lower knee).
- 17:03 First C1 launch: run-case.sh bug (bash `pids[$arm]` unbound under set -u) launched only v1 → solo v1 rep archived under C1/S1/solo-v1-rep (not used).
- 17:17 Second launch aborted: gamebench bug — worker RNG deterministic across runs → purchase request_id UUID collisions (mysql_1062, 0.5% errors in solo run). Fixed (per-run nonce mixed into worker seed), binary redeployed; archived under C1/S1/aborted-uuid-bug.
- 17:19 C1/S1 @5000/s, 3 reps × 600 s, both arms concurrent — RUNNING. Note: `accounts` passed to runner = 2,294,672 (information_schema estimate from load json) on both arms — a consistent subset of the 2.5M key space.
- 17:05 Exp A deployed by user (`deploy-expA.ps1`): expa-pg (D8ds_v5, PG 17.10, SameZone HA, P15 256 GiB), expa-hz (HorizonDB 8 vCore, replicaCount 1, PG 17.9), expa-bench-vm (D16ds_v5, australiaeast). Dataset loaded on both: 23.06 GiB. shared_buffers: PG 8 GB vs HorizonDB 45,058 MB → G1 fails for HorizonDB (dataset 0.51× its buffer cache) — documented confound; HZ node evidently has far more memory per vCore.
