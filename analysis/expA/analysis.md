# v1 vs v2 paired analysis

Generated 2026-08-17T18:36:06.123585+00:00

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
