# Run log (UTC)

## 2026-08-17

- 15:59 Reused servers `mysqlbm-euson-v1` (E8ds_v5, Premium_LRS) / `-v2` (E8ds_v6, PremiumV2_LRS). Set `innodb_buffer_pool_size`=8 GiB on both (pre-registered control), storage 64→128 GiB, v1 IOPS 492→5000 (same-IOPS cell).
- 16:07 v1 HA enable failed: `AutoGrowDisabledNotSupportedForHA` → enabled autogrow on v1, retried HA (SameZone). v2 HA SameZone enabled OK (autoGrow reported Disabled but accepted).
- 16:03 Bench VM `mysqlbm-euson-lg-vm` resized D4ds_v5→D16ds_v5, public IP + NSG (SSH/3000 from operator IP only). Key Vault (private, RBAC) not readable by operator → MySQL password must be placed on the VM by the user (`scripts/vm-fetch-mysql-secret.sh` via run-command).
- 16:15 Monitoring VM `bench-mon-vm` (D4s_v5) created; InfluxDB2 + Grafana + Telegraf installed via `infra/monitoring/cloud-init.yaml` (executed as setup.sh). MySQL inputs collected from bench VM telegraf (credentials live there), pushed to monitoring InfluxDB.
- Time budget from user: ~7h total; scope reduced to C1 (8 vCore/SameZone HA) and C5 (16 vCore/SameZone HA), 3 reps × 10 min, arms run concurrently.
