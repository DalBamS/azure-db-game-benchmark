#!/bin/bash
# Closed-loop concurrency staircase on both arms concurrently -> knee JSON per arm.
# usage: knee-both.sh CASE SCENARIO [steps] [step-sec]
source "$(dirname "$0")/common.sh"
CASE="$1"; SCEN="${2:-S1}"; STEPS="${3:-32,64,128,256,384,512}"; STEPSEC="${4:-90}"
ACCOUNTS=$(python3 -c "import json,glob;print(sorted(glob.glob('$RESULTS/load-*-v1.json'))[-1])" | xargs -I{} python3 -c "import json;print(json.load(open('{}'))['approx_rows'].get('accounts',0))")
ACCOUNTS=${ACCOUNTS_OVERRIDE:-$ACCOUNTS}
STAMP=$(ts); OUT="$RESULTS/$CASE/$SCEN"; mkdir -p "$OUT"
for arm in v1 v2; do
  dsnvar="DSN_$(echo $arm | tr a-z A-Z)"
  "$GB" knee -dsn "${!dsnvar}" -label "$arm" -scenario "$SCEN" -steps "$STEPS" -step-sec "$STEPSEC" -accounts "$ACCOUNTS" -slots "${SLOTS:-16}" -nic-bps "$NIC_BPS" -out "$OUT/knee-$STAMP-$arm.json" > "$OUT/knee-$STAMP-$arm.log" 2>&1 &
done
wait
for arm in v1 v2; do
  python3 -c "import json;d=json.load(open('$OUT/knee-$STAMP-$arm.json'));print('$arm knee_workers=%d knee_tps=%.0f rate65=%.0f'%(d['knee_workers'],d['knee_tps'],d['recommended_rate_65pct'])); [print('   ',s) for s in d['steps']]"
done
