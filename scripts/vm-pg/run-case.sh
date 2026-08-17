#!/bin/bash
source "$(dirname "$0")/common.sh"
CASE="$1"; SCEN="$2"; RATE="$3"; REPS="${4:-3}"; MEASURE="${5:-600}"; WORKERS="${6:-256}"; WARMUP="${7:-120}"; ACCOUNTS="${ACCOUNTS_OVERRIDE:-2500000}"
OUT="$RESULTS/$CASE/$SCEN/rate$RATE"; mkdir -p "$OUT"
echo "$(ts) case=$CASE scen=$SCEN rate=$RATE reps=$REPS measure=$MEASURE workers=$WORKERS" | tee -a "$OUT/run-case.log"
declare -A pids
for rep in $(seq 1 "$REPS"); do
  STAMP=$(ts); echo "$(ts) rep $rep start" | tee -a "$OUT/run-case.log"
  for arm in $ARMS; do
    PG_DSN="$(dsn_of $arm)" "$GB" run -driver pgx -label "$arm" -scenario "$SCEN" -mode open -rate "$RATE" -workers "$WORKERS" -warmup "$WARMUP" -warmup-max 480 -measure "$MEASURE" \
      -accounts "$ACCOUNTS" -slots "${SLOTS:-20}" -nic-bps "$NIC_BPS" -out "$OUT/rep$rep-$STAMP-$arm.json" > "$OUT/rep$rep-$STAMP-$arm.log" 2>&1 &
    pids[$arm]=$!
  done
  for arm in $ARMS; do wait ${pids[$arm]} || echo "$(ts) rep $rep $arm exit=$? (gates may have failed)" | tee -a "$OUT/run-case.log"; done
  for arm in $ARMS; do tail -n 14 "$OUT/rep$rep-$STAMP-$arm.log" | grep -E "done:|gate|server:" | sed "s/^/  [$arm] /" | tee -a "$OUT/run-case.log"; PG_DSN="$(dsn_of $arm)" "$GB" check -driver pgx -out "$OUT/rep$rep-$STAMP-$arm.invariants.json" 2>/dev/null; done
  echo "$(ts) rep $rep done" | tee -a "$OUT/run-case.log"; [ "$rep" -lt "$REPS" ] && sleep 60
done
echo "$(ts) case complete: $OUT" | tee -a "$OUT/run-case.log"
