#!/bin/bash
# Create schema and load the dataset on both arms concurrently.
# usage: load-both.sh [accounts] [slots] [profile-bytes] [attrs-bytes]
source "$(dirname "$0")/common.sh"
ACCOUNTS="${1:-2500000}"; SLOTS="${2:-16}"; PROF="${3:-1024}"; ATTRS="${4:-320}"
STAMP=$(ts)
for arm in v1 v2; do
  dsnvar="DSN_$(echo $arm | tr a-z A-Z)"
  (
    "$GB" schema -dsn "${!dsnvar}" && "$GB" reset -dsn "${!dsnvar}" &&
    "$GB" load -dsn "${!dsnvar}" -accounts "$ACCOUNTS" -slots "$SLOTS" -profile-bytes "$PROF" -attrs-bytes "$ATTRS" -load-workers 24 -batch 400 -out "$RESULTS/load-$STAMP-$arm.json"
  ) > "$RESULTS/load-$STAMP-$arm.log" 2>&1 &
done
wait
echo "load done: $RESULTS/load-$STAMP-*.json"
for arm in v1 v2; do
  python3 -c "import json;d=json.load(open('$RESULTS/load-$STAMP-$arm.json'));print('$arm total_gib=%.2f'%d['total_gib'], {k:round(v,2) for k,v in d['table_gib'].items() if v>0.01})"
done
