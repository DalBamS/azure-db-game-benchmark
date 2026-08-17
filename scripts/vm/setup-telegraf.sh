#!/bin/bash
# Render telegraf config on the bench VM from ~/.bench/mysql.env + ~/.bench/monitoring.env and start it.
set -euo pipefail
set -a; source "$HOME/.bench/mysql.env"; source "$HOME/.bench/monitoring.env"; set +a
export V1_HOST="${V1_HOST:-mysqlbm-euson-v1.mysql.database.azure.com}" V2_HOST="${V2_HOST:-mysqlbm-euson-v2.mysql.database.azure.com}"
sudo apt-get install -y -qq gettext-base >/dev/null 2>&1 || true
envsubst < "$HOME/scripts/telegraf-bench.conf.tmpl" | sudo tee /etc/telegraf/telegraf.conf >/dev/null
sudo systemctl enable --now telegraf >/dev/null 2>&1; sudo systemctl restart telegraf
sleep 3; systemctl is-active telegraf; sudo journalctl -u telegraf --no-pager -n 5 | tail -5
