#!/bin/bash
# Experiment A (PostgreSQL Flexible vs HorizonDB) VM-side environment. Arms: postgres, horizon.
set -euo pipefail
RESULTS="${RESULTS:-$HOME/results}"; mkdir -p "$RESULTS"
# shellcheck disable=SC1090
source "$HOME/.bench/pg.env"
DBNAME="${DBNAME:-benchmark}"
enc() { python3 -c "import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1],safe=''))" "$1"; }
PW_ENC=$(enc "$PG_PASSWORD")
export DSN_POSTGRES="postgres://${PG_USER}:${PW_ENC}@${PG_HOST}:5432/${DBNAME}?sslmode=require&connect_timeout=10&default_query_exec_mode=cache_statement"
export DSN_HORIZON="postgres://${PG_USER}:${PW_ENC}@${HZ_HOST}:5432/${DBNAME}?sslmode=require&connect_timeout=10&default_query_exec_mode=cache_statement"
GB="$HOME/gamebench"
NIC_BPS="${NIC_BPS:-1562500000}"
ARMS="postgres horizon"
dsn_of() { case "$1" in postgres) echo "$DSN_POSTGRES";; horizon) echo "$DSN_HORIZON";; esac; }
ts() { date -u +%Y%m%dT%H%M%SZ; }
