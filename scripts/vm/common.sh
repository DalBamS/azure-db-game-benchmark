#!/bin/bash
# Shared environment for VM-side scripts. Credentials are read from ~/.bench/mysql.env
# (never committed, never printed).
set -euo pipefail
BENCH_HOME="${BENCH_HOME:-$HOME}"
RESULTS="${RESULTS:-$HOME/results}"
mkdir -p "$RESULTS"
# shellcheck disable=SC1090
source "$HOME/.bench/mysql.env"
V1_HOST="${V1_HOST:-mysqlbm-euson-v1.mysql.database.azure.com}"
V2_HOST="${V2_HOST:-mysqlbm-euson-v2.mysql.database.azure.com}"
DBNAME="${DBNAME:-benchmark}"
dsn() { # $1 = host
  echo "${MYSQL_USER}:${MYSQL_PASSWORD}@tcp($1:3306)/${DBNAME}?tls=true&interpolateParams=true&parseTime=true&timeout=10s&readTimeout=60s&writeTimeout=60s"
}
export DSN_V1; DSN_V1="$(dsn "$V1_HOST")"
export DSN_V2; DSN_V2="$(dsn "$V2_HOST")"
GB="$BENCH_HOME/gamebench"
# NIC limit for G3: D16ds_v5 ~ 12.5 Gbps => 1.5625e9 B/s
NIC_BPS="${NIC_BPS:-1562500000}"
ts() { date -u +%Y%m%dT%H%M%SZ; }
