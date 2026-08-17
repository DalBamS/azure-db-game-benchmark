#!/bin/bash
# Runs ON the bench VM (via az vm run-command). Reads the MySQL admin password from Key Vault
# using the VM's managed identity (private endpoint) and stores it ONLY in the VM-local file
# /home/benchadmin/.bench/mysql.env. The value is never printed.
set -e
TOKEN=$(curl -s -H Metadata:true 'http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fvault.azure.net' | python3 -c 'import sys,json;print(json.load(sys.stdin)["access_token"])')
curl -s -H "Authorization: Bearer $TOKEN" 'https://kv-mysqlbm-euson.vault.azure.net/secrets/mysql-administrator-password?api-version=7.4' > /tmp/kv.json
mkdir -p /home/benchadmin/.bench
python3 - <<'PY'
import json
d=json.load(open('/tmp/kv.json'))
if 'value' in d:
    open('/home/benchadmin/.bench/mysql.env','w').write('MYSQL_USER=mysqladmin\nMYSQL_PASSWORD=%s\n' % d['value'])
    print('ok')
else:
    print('ERROR', d)
PY
rm -f /tmp/kv.json
chown -R benchadmin:benchadmin /home/benchadmin/.bench
chmod 600 /home/benchadmin/.bench/mysql.env 2>/dev/null || true
