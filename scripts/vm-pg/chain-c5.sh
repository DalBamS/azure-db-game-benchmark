#!/bin/bash
# Exp A chain (PG/HZ): C5 S1 reps 4-5, S2 x3, S4 x3, S3 x3 (burst)
export SLOTS=20
cd ~
REP_START=4 bash ~/scripts/run-case.sh C5 S1 5500 5 600 256 120
bash ~/scripts/run-case.sh C5 S2 3300 3 600 256 120
bash ~/scripts/run-case.sh C5 S4 3300 3 600 256 120
bash ~/scripts/run-case.sh C5 S3 2200 3 600 256 120 -burst-at 240 -burst-sec 120 -burst-factor 3
echo "$(date -u) CHAIN_DONE"
