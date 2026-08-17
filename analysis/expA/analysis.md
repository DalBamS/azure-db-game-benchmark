# v1 vs v2 paired analysis

Generated 2026-08-17T19:56:03.719984+00:00

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

Repetition gate summary:
- rep 1 postgres: ok=True tps=5500 p99=27391us azmon=True inv=True
- rep 1 horizon: ok=True tps=5500 p99=4379us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=True tps=5500 p99=31935us azmon=True inv=True
- rep 2 horizon: ok=True tps=5500 p99=4407us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=True tps=5500 p99=87871us azmon=True inv=True
- rep 3 horizon: ok=True tps=5500 p99=4235us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

## C5 / S1 / 5500 arrivals/s  (postgres vs horizon)

pairs: 1 usable of 3 (hard-gate-failed pairs excluded; soft gates: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])

| Outcome | n | postgres median | horizon median | postgres CV% | horizon CV% | Δ% (horizon vs postgres, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 1 | 5,500.1 | 5,500.0 | - | - | - | - | insufficient pairs for effect/CI |
| Latency p50 (us) | 1 | 576.0 | 473.0 | - | - | - | - | insufficient pairs for effect/CI |
| Latency p95 (us) | 1 | 15,207.0 | 3,287.0 | - | - | - | - | insufficient pairs for effect/CI |
| Latency p99 (us) | 1 | 55,007.0 | 3,691.0 | - | - | - | - | insufficient pairs for effect/CI |
| Latency p99.9 (us) | 1 | 106,687.0 | 4,267.0 | - | - | - | - | insufficient pairs for effect/CI |
| Error rate | 1 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 1 | 1,285.4 | 0.6017 | - | - | - | - |  |
| Server write IOPS | 1 | 1,748.2 | 2,433.9 | - | - | - | - |  |
| Buffer pool hit ratio | 1 | 0.9884 | 1.0000 | - | - | - | - |  |
| Service time p99 (us, excl. queue) | 1 | 54,879.0 | 3,541.0 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 postgres: ok=True tps=5500 p99=55007us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 1 horizon: ok=True tps=5500 p99=3691us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 2 postgres: ok=False tps=5405 p99=15613951us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 2 horizon: ok=True tps=5500 p99=3701us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
- rep 3 postgres: ok=False tps=5500 p99=3864575us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G4_queue_delay'] (soft-gate warnings: ['G1_dataset_vs_bufferpool'])
- rep 3 horizon: ok=True tps=5500 p99=3683us azmon=True inv=True FAILED: ['G1_dataset_vs_bufferpool', 'G2_storage_used'] (soft-gate warnings: ['G1_dataset_vs_bufferpool', 'G2_storage_used'])
