#!/bin/bash
# Main measurement: N repetitions; in each repetition both arms run concurrently
# (same wall-clock window => identical client/network noise; G3 verifies client headroom).
# usage: run-case.sh CASE SCENARIO RATE REPS [measure-sec] [workers] [warmup-sec] [extra gamebench flags...]
source "$(dirname "$0")/common.sh"
CASE="$1"; SCEN="$2"; RATE="$3"; REPS="${4:-3}"; MEASURE="${5:-600}"; WORKERS="${6:-256}"; WARMUP="${7:-120}"; shift 7 || shift $#
EXTRA="$*"
ACCOUNTS=$(python3 -c "import json,glob;f=sorted(glob.glob('$RESULTS/load-*-v1.json'))[-1];print(json.load(open(f))['approx_rows'].get('accounts',0))")
ACCOUNTS=${ACCOUNTS_OVERRIDE:-$ACCOUNTS}
OUT="$RESULTS/$CASE/$SCEN/rate$RATE"; mkdir -p "$OUT"
echo "$(ts) case=$CASE scen=$SCEN rate=$RATE reps=$REPS measure=$MEASURE workers=$WORKERS warmup=$WARMUP accounts=$ACCOUNTS extra=$EXTRA" | tee -a "$OUT/run-case.log"
for rep in $(seq 1 "$REPS"); do
  STAMP=$(ts)
  echo "$(ts) rep $rep start" | tee -a "$OUT/run-case.log"
  for arm in v1 v2; do
    dsnvar="DSN_$(echo $arm | tr a-z A-Z)"
    "$GB" run -dsn "${!dsnvar}" -label "$arm" -scenario "$SCEN" -mode open -rate "$RATE" -workers "$WORKERS" -warmup "$WARMUP" -warmup-max 480 -measure "$MEASURE" \
      -accounts "$ACCOUNTS" -slots "${SLOTS:-16}" -nic-bps "$NIC_BPS" $EXTRA -out "$OUT/rep$rep-$STAMP-$arm.json" > "$OUT/rep$rep-$STAMP-$arm.log" 2>&1 &
    pids[$arm]=$!
  done
  wait ${pids[v1]} || echo "$(ts) rep $rep v1 exit=$? (gates may have failed)" | tee -a "$OUT/run-case.log"
  wait ${pids[v2]} || echo "$(ts) rep $rep v2 exit=$? (gates may have failed)" | tee -a "$OUT/run-case.log"
  for arm in v1 v2; do
    tail -n 14 "$OUT/rep$rep-$STAMP-$arm.log" | grep -E "done:|gate|server:" | sed "s/^/  [$arm] /" | tee -a "$OUT/run-case.log"
  done
  # invariants after each rep (G9)
  for arm in v1 v2; do
    dsnvar="DSN_$(echo $arm | tr a-z A-Z)"
    "$GB" check -dsn "${!dsnvar}" -out "$OUT/rep$rep-$STAMP-$arm.invariants.json" 2>/dev/null
  done
  echo "$(ts) rep $rep done; cooldown 60s" | tee -a "$OUT/run-case.log"
  [ "$rep" -lt "$REPS" ] && sleep 60
done
echo "$(ts) case complete: $OUT" | tee -a "$OUT/run-case.log"
