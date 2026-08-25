# v1 vs v2 paired analysis

Generated 2026-08-25T22:10:41.571766+00:00

## C1 / S1 / 5500 arrivals/s  (postgres vs horizon)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 5,499.8 | 5,500.0 | 0.0032 | 0.0000 | +0.00% | [+0.00%, +0.01%] | horizon higher (better); n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 728.0 | 551.0 | 0.7793 | 0.1815 | -24.48% | [-25.27%, -23.86%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 14,239.0 | 3,617.0 | 40.3 | 0.4481 | -78.99% | [-86.41%, -73.12%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 31,935.0 | 4,379.0 | 68.6 | 2.1 | -89.79% | [-95.18%, -84.01%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 142,463.0 | 9,815.0 | 54.2 | 6.9 | -92.77% | [-96.17%, -85.69%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 3,087.5 | 0.0000 | 4.6 | - | - | - |  |
| Server write IOPS | 3 | 1,490.3 | 2,639.7 | 8.0 | 35.1 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9730 | 1.0 | 0.1580 | 0.0000 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 29,455.0 | 4,215.0 | 60.1 | 2.0 | -89.19% | [-94.40%, -84.30%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=True tps=5500 p99=27391us azmon=True inv=True
- rep 1 horizon: ok=True tps=5500 p99=4379us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=True tps=5500 p99=31935us azmon=True inv=True
- rep 2 horizon: ok=True tps=5500 p99=4407us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=True tps=5500 p99=87871us azmon=True inv=True
- rep 3 horizon: ok=True tps=5500 p99=4235us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

## C5 / S1 / 5500 arrivals/s  (postgres vs horizon)

pairs: 2 usable of 5 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 2 | 5,500.0 | 5,500.0 | 0.0008 | 0.0000 | -0.00% | [-0.00%, -0.00%] | horizon lower (worse); n=2<5 (below protocol baseline) |
| Latency p50 (us) | 2 | 570.5 | 474.0 | 1.4 | 0.2984 | -16.91% | [-17.88%, -15.93%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Latency p95 (us) | 2 | 12,759.0 | 3,291.0 | 27.1 | 0.1719 | -73.72% | [-78.38%, -68.04%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Latency p99 (us) | 2 | 34,971.0 | 3,695.0 | 81.0 | 0.1531 | -87.11% | [-93.29%, -75.23%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Latency p99.9 (us) | 2 | 73,839.0 | 4,287.0 | 62.9 | 0.6598 | -93.52% | [-96.00%, -89.49%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Error rate | 2 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 2 | 1,288.1 | 0.3008 | 0.2944 | 141.4 | - | - |  |
| Server write IOPS | 2 | 1,155.9 | 2,205.8 | 72.5 | 14.6 | - | - |  |
| Buffer pool hit ratio | 2 | 0.9883 | 1.0000 | 0.0129 | 0.0004 | - | - |  |
| Service time p99 (us, excl. queue) | 2 | 34,819.0 | 3,543.0 | 81.5 | 0.0798 | -87.55% | [-93.55%, -75.98%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 2 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=True tps=5500 p99=55007us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 1 horizon: ok=True tps=5500 p99=3691us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=False tps=5405 p99=15613951us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 2 horizon: ok=True tps=5500 p99=3701us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=False tps=5500 p99=3864575us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G4_queue_delay'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 3 horizon: ok=True tps=5500 p99=3683us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 4 postgres: ok=True tps=5500 p99=14935us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 4 horizon: ok=True tps=5500 p99=3699us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 5 postgres: ok=False tps=5500 p99=1354751us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G4_queue_delay'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 5 horizon: ok=True tps=5500 p99=3707us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

## C5 / S2 / 3300 arrivals/s  (postgres vs horizon)

pairs: 0 usable of 3 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p50 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p95 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p99 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p99.9 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Error rate | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 0 | - | - | - | - | - | - |  |
| Server write IOPS | 0 | - | - | - | - | - | - |  |
| Buffer pool hit ratio | 0 | - | - | - | - | - | - |  |
| Service time p99 (us, excl. queue) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=False tps=3300 p99=371711us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used', 'G4_queue_delay'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 1 horizon: ok=True tps=3300 p99=3965us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=False tps=3263 p99=11665407us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used', 'G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 horizon: ok=True tps=3300 p99=3981us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=False tps=3194 p99=20299775us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used', 'G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 horizon: ok=True tps=3300 p99=3953us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

## C5 / S3 / 2200 arrivals/s  (postgres vs horizon)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 3,080.0 | 3,080.0 | 0.0016 | 0.0001 | +0.00% | [+0.00%, +0.00%] | horizon higher (better); n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 6,051.0 | 2,375.0 | 3.6 | 0.7650 | -60.53% | [-62.09%, -58.90%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 10,743.0 | 3,405.0 | 43.6 | 0.2117 | -73.10% | [-83.12%, -63.74%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 16,799.0 | 3,811.0 | 167.9 | 0.2186 | -94.48% | [-99.74%, -71.76%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 303,615.0 | 4,375.0 | 137.2 | 0.2643 | -98.77% | [-99.81%, -93.35%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 1,557.4 | 0.0000 | 3.8 | 173.2 | - | - |  |
| Server write IOPS | 3 | 586.6 | 1,025.9 | 33.7 | 0.8615 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9624 | 1.0 | 0.1243 | 0.0014 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 16,399.0 | 3,477.0 | 97.5 | 0.1521 | -85.86% | [-95.10%, -72.90%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 3 | 0.0000 | 0.0000 | 173.2 | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=True tps=3080 p99=1449983us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G4_queue_delay'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G4_queue_delay'])
- rep 1 horizon: ok=True tps=3080 p99=3799us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=True tps=3080 p99=16799us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 2 horizon: ok=True tps=3080 p99=3815us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=True tps=3080 p99=13495us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 3 horizon: ok=True tps=3080 p99=3811us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

## C7 / S1 / 5500 arrivals/s  (postgres vs horizon)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 5,500.0 | 5,500.0 | 0.0001 | 0.0000 | +0.00% | [+0.00%, +0.00%] | horizon higher (better); n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 563.0 | 479.0 | 0.3552 | 0.7327 | -14.86% | [-15.45%, -14.51%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 5,455.0 | 3,329.0 | 6.4 | 0.4266 | -39.31% | [-43.45%, -35.42%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 7,655.0 | 3,727.0 | 124.0 | 0.1421 | -75.04% | [-94.05%, -46.35%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 19,503.0 | 4,307.0 | 131.7 | 0.2982 | -88.45% | [-97.61%, -70.73%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 1,375.0 | 0.0017 | 1.3 | 172.6 | - | - |  |
| Server write IOPS | 3 | 1,103.1 | 2,063.9 | 27.9 | 1.9 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9874 | 1.0000 | 0.0157 | 0.0004 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 7,551.0 | 3,577.0 | 97.0 | 0.1480 | -70.53% | [-89.69%, -47.64%] | horizon lower (better); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=True tps=5500 p99=6947us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 1 horizon: ok=True tps=5500 p99=3727us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=True tps=5500 p99=7655us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 2 horizon: ok=True tps=5500 p99=3729us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=True tps=5500 p99=62495us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 3 horizon: ok=True tps=5500 p99=3719us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

## C7 / S1 / 5501 arrivals/s  (postgres vs horizon)

pairs: 0 usable of 3 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p50 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p95 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p99 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Latency p99.9 (us) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Error rate | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 0 | - | - | - | - | - | - |  |
| Server write IOPS | 0 | - | - | - | - | - | - |  |
| Buffer pool hit ratio | 0 | - | - | - | - | - | - |  |
| Service time p99 (us, excl. queue) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=False tps=3567 p99=35160063us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors', 'G7_steady_state']
- rep 1 horizon: ok=True tps=5501 p99=3867us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=False tps=5087 p99=16924671us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors']
- rep 2 horizon: ok=True tps=5501 p99=3889us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=False tps=5409 p99=8896511us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors']
- rep 3 horizon: ok=True tps=5501 p99=3953us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

## C7 / S4 / 3300 arrivals/s  (postgres vs horizon)

pairs: 2 usable of 3 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 2 | 3,300.0 | 3,300.0 | 0.0002 | 0.0002 | +0.00% | [+0.00%, +0.00%] | horizon higher (better); n=2<5 (below protocol baseline) |
| Latency p50 (us) | 2 | 3,781.0 | 2,608.0 | 0.6733 | 0.1627 | -31.02% | [-31.27%, -30.77%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Latency p95 (us) | 2 | 6,989.0 | 3,591.0 | 11.5 | 0.2363 | -48.45% | [-52.55%, -44.00%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Latency p99 (us) | 2 | 20,427.0 | 3,967.0 | 77.6 | 0.4278 | -76.77% | [-87.50%, -56.86%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Latency p99.9 (us) | 2 | 48,759.0 | 5,169.0 | 63.3 | 2.5 | -88.15% | [-92.81%, -80.47%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Error rate | 2 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 2 | 358.8 | 0.3308 | 3.7 | 141.4 | - | - |  |
| Server write IOPS | 2 | 836.7 | 1,642.4 | 44.4 | 3.7 | - | - |  |
| Buffer pool hit ratio | 2 | 0.9945 | 1.0000 | 0.0199 | 0.0007 | - | - |  |
| Service time p99 (us, excl. queue) | 2 | 20,219.0 | 3,750.0 | 77.8 | 0.2640 | -77.79% | [-88.06%, -58.69%] | horizon lower (better); n=2<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 2 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=True tps=3300 p99=31631us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 1 horizon: ok=True tps=3300 p99=3955us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=True tps=3300 p99=9223us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 horizon: ok=True tps=3300 p99=3979us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=False tps=3300 p99=3463167us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used', 'G4_queue_delay'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 horizon: ok=True tps=3300 p99=3991us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
