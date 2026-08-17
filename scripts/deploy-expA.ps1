# Run this in YOUR PowerShell (interactive). It prompts for the DB admin password locally and passes it
# only to the ARM deployment; the password also lands in the bench VM's ~/.bench/pg.env via cloud-init.
# Nothing is printed or stored on this machine.
param(
  [string]$Rg = 'rg-expa-pg-hz',
  [string]$OperatorIp = '121.143.198.178',
  [string]$SshPubKeyPath = "$env:USERPROFILE\.ssh\bench2_ed25519.pub"
)
az account set -s 'Euson Internal Subscription'
$sec = Read-Host -Prompt 'DB admin password for expa-pg / expa-hz (min 8 chars, upper+lower+digit)' -AsSecureString
$plain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($sec))
$key = (Get-Content $SshPubKeyPath -Raw).Trim()
az deployment group create -g $Rg -n expa-c1 -f "$PSScriptRoot\..\infra\expA\main.bicep" `
  -p administratorLoginPassword="$plain" benchAdminSshKey="$key" operatorIp=$OperatorIp `
  --query "{state:properties.provisioningState,outputs:properties.outputs}" -o json
$plain = $null
