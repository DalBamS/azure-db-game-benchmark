#!/bin/bash
source "$(dirname "$0")/common.sh"
ACCOUNTS="${1:-2500000}"; SLOTS="${2:-20}"; PROF="${3:-1024}"; ATTRS="${4:-320}"
STAMP=$(ts)
for arm in $ARMS; do
  ( PG_DSN="$(dsn_of $arm)" "$GB" schema -driver pgx && PG_DSN="$(dsn_of $arm)" "$GB" reset -driver pgx &&
    PG_DSN="$(dsn_of $arm)" "$GB" load -driver pgx -accounts "$ACCOUNTS" -slots "$SLOTS" -profile-bytes "$PROF" -attrs-bytes "$ATTRS" -load-workers 24 -batch 400 -out "$RESULTS/load-$STAMP-$arm.json"
  ) > "$RESULTS/load-$STAMP-$arm.log" 2>&1 &
done
wait
echo "load done: $RESULTS/load-$STAMP-*.json"
for arm in $ARMS; do python3 -c "import json;d=json.load(open('$RESULTS/load-$STAMP-$arm.json'));print('$arm total_gib=%.2f'%d['total_gib'], 'shared_buffers=',d['variables'].get('shared_buffers'))"; done
