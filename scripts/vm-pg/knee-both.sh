#!/bin/bash
source "$(dirname "$0")/common.sh"
CASE="$1"; SCEN="${2:-S1}"; STEPS="${3:-32,64,128,256,384}"; STEPSEC="${4:-90}"; ACCOUNTS="${ACCOUNTS_OVERRIDE:-2500000}"
STAMP=$(ts); OUT="$RESULTS/$CASE/$SCEN"; mkdir -p "$OUT"
for arm in $ARMS; do
  PG_DSN="$(dsn_of $arm)" "$GB" knee -driver pgx -label "$arm" -scenario "$SCEN" -steps "$STEPS" -step-sec "$STEPSEC" -accounts "$ACCOUNTS" -slots "${SLOTS:-20}" -nic-bps "$NIC_BPS" -out "$OUT/knee-$STAMP-$arm.json" > "$OUT/knee-$STAMP-$arm.log" 2>&1 &
done
wait
for arm in $ARMS; do python3 -c "import json;d=json.load(open('$OUT/knee-$STAMP-$arm.json'));print('$arm knee_workers=%d knee_tps=%.0f rate65=%.0f'%(d['knee_workers'],d['knee_tps'],d['recommended_rate_65pct'])); [print('   ',s) for s in d['steps']]"; done
