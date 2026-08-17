#!/bin/bash
# Exp B chain (MySQL): C5 S1 reps 4-5, S2 x3, S4 x3, S3 x3 (burst)
export SLOTS=20 SKIP_CHECK=1
cd ~
REP_START=4 bash ~/scripts/run-case.sh C5 S1 5000 5 600 256 120
bash ~/scripts/run-case.sh C5 S2 3000 3 600 256 120
bash ~/scripts/run-case.sh C5 S4 3000 3 600 256 120
bash ~/scripts/run-case.sh C5 S3 2000 3 600 256 120 -burst-at 240 -burst-sec 120 -burst-factor 3
echo "$(date -u) CHAIN_DONE"
