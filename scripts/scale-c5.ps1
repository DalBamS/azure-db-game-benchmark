# Scale both experiments' arms from 8 vCore (C1) to 16 vCore (C5). Idempotent.
param([switch]$MySQL, [switch]$PgHz)
az account set -s 'Euson Internal Subscription'
$log = Join-Path $PSScriptRoot '..\results\scale-c5.log'
function Log($m) { $line = "$(Get-Date -Format o) $m"; Write-Host $line; Add-Content -Path $log -Value $line }
if ($MySQL) {
  Log 'MySQL v1 -> Standard_E16ds_v5'
  Start-Job -ScriptBlock { az mysql flexible-server update -g mysql-storage-benchmark -n mysqlbm-euson-v1 --sku-name Standard_E16ds_v5 --tier MemoryOptimized -o json 2>&1 } | Out-Null
  Log 'MySQL v2 -> Standard_E16ds_v6'
  Start-Job -ScriptBlock { az mysql flexible-server update -g mysql-storage-benchmark -n mysqlbm-euson-v2 --sku-name Standard_E16ds_v6 --tier MemoryOptimized -o json 2>&1 } | Out-Null
}
if ($PgHz) {
  Log 'PG -> Standard_D16ds_v5'
  Start-Job -ScriptBlock { az postgres flexible-server update -g rg-expa-pg-hz -n expa-pg --sku-name Standard_D16ds_v5 --tier GeneralPurpose -o json 2>&1 } | Out-Null
  Log 'HorizonDB -> vCores 16'
  Start-Job -ScriptBlock { az rest --method patch --url "/subscriptions/7784c8b4-64ba-4b09-bee8-ee6f8f9a7309/resourceGroups/rg-expa-pg-hz/providers/Microsoft.HorizonDB/clusters/expa-hz?api-version=2026-01-20-preview" --body '{"properties":{"vCores":16}}' -o json 2>&1 } | Out-Null
}
Get-Job | Wait-Job | Receive-Job | ForEach-Object { Log ($_ | Out-String).Trim() }
Log 'scale done'
