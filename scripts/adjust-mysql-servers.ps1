# Adjust the reused MySQL servers for Phase 0 (C1 = 8 vCore / Same-zone HA / no replica).
# Steps are idempotent; each prints JSON of the resulting state.
param(
  [string]$Rg = 'mysql-storage-benchmark',
  [string]$BufferPoolBytes = '8589934592',   # 8 GiB, pinned identically on both arms (pre-registered control)
  [int]$StorageGB = 128,
  [int]$Iops = 5000,
  [string]$Ha = 'SameZone'
)
$ErrorActionPreference = 'Continue'
az account set -s 'Euson Internal Subscription'
$log = Join-Path $PSScriptRoot '..\results\infra-adjust.log'
function Log($m) { $line = "$(Get-Date -Format o) $m"; Write-Host $line; Add-Content -Path $log -Value $line }

foreach ($s in 'mysqlbm-euson-v1','mysqlbm-euson-v2') {
  Log "[$s] set innodb_buffer_pool_size=$BufferPoolBytes"
  az mysql flexible-server parameter set -g $Rg -s $s --name innodb_buffer_pool_size --value $BufferPoolBytes -o none
  Log "[$s] storage-size=$StorageGB"
  az mysql flexible-server update -g $Rg -n $s --storage-size $StorageGB -o none
}
Log "[v1] iops=$Iops (same-IOPS cell)"
az mysql flexible-server update -g $Rg -n mysqlbm-euson-v1 --iops $Iops -o none

foreach ($s in 'mysqlbm-euson-v1','mysqlbm-euson-v2') {
  Log "[$s] enable HA=$Ha"
  az mysql flexible-server update -g $Rg -n $s --high-availability $Ha -o none 2>&1 | ForEach-Object { Log "[$s] $_" }
  az mysql flexible-server show -g $Rg -n $s --query "{name:name,sku:sku.name,ha:highAvailability,storage:storage,state:state}" -o json | ForEach-Object { Log $_ }
}
Log 'done'
