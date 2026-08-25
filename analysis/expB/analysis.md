# v1 vs v2 paired analysis

Generated 2026-08-25T22:10:33.876097+00:00

## C1 / S1 / 5000 arrivals/s  (v1 vs v2)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 5,000.0 | 5,000.0 | 0.0005 | 0.0005 | -0.00% | [-0.00%, +0.00%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 3,539.0 | 2,261.0 | 6.2 | 18.6 | -28.64% | [-39.39%, -13.51%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 19,423.0 | 15,263.0 | 4.0 | 69.2 | +11.33% | [-25.24%, +132.96%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 25,199.0 | 18,511.0 | 20.0 | 79.8 | +2.82% | [-44.70%, +156.39%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 43,135.0 | 26,303.0 | 72.8 | 77.8 | -32.94% | [-79.33%, +108.61%] | CI includes 0; n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 2,225.3 | 1,518.8 | 22.4 | 36.4 | - | - |  |
| Server write IOPS | 3 | 2,183.1 | 3,299.1 | 3.9 | 21.4 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9752 | 0.9791 | 0.5177 | 0.6284 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 25,039.0 | 18,351.0 | 17.9 | 80.2 | +3.98% | [-42.85%, +157.39%] | CI includes 0; n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 v1: ok=True tps=5000 p99=23167us azmon=True inv=True
- rep 1 v2: ok=True tps=5000 p99=17759us azmon=True inv=True
- rep 2 v1: ok=True tps=5000 p99=33471us azmon=True inv=True
- rep 2 v2: ok=True tps=5000 p99=18511us azmon=True inv=True
- rep 3 v1: ok=True tps=5000 p99=25199us azmon=True inv=True
- rep 3 v2: ok=True tps=5000 p99=64607us azmon=True inv=True

## C3 / S1 / 5000 arrivals/s  (v1 vs v2)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 5,000.0 | 5,000.0 | 0.0002 | 0.0018 | +0.00% | [-0.00%, +0.00%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 3,711.0 | 2,573.0 | 4.9 | 0.8794 | -32.09% | [-36.19%, -29.20%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 20,799.0 | 28,831.0 | 2.0 | 13.6 | +30.16% | [+8.66%, +42.46%] | v2 higher (worse); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 27,455.0 | 40,639.0 | 6.6 | 14.0 | +41.95% | [+14.25%, +64.09%] | v2 higher (worse); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 79,679.0 | 55,167.0 | 73.0 | 12.8 | -44.62% | [-72.21%, +9.81%] | CI includes 0; n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 1,720.9 | 1,730.7 | 3.6 | 0.8464 | - | - |  |
| Server write IOPS | 3 | 2,347.5 | 3,095.3 | 0.2502 | 0.5912 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9766 | 0.9765 | 0.0772 | 0.0234 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 26,703.0 | 40,447.0 | 6.3 | 14.0 | +43.59% | [+14.97%, +64.80%] | v2 higher (worse); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 v1: ok=True tps=5000 p99=28079us azmon=True inv=True
- rep 1 v2: ok=True tps=5000 p99=32079us azmon=True inv=True
- rep 2 v1: ok=True tps=5000 p99=24767us azmon=True inv=True
- rep 2 v2: ok=True tps=5000 p99=40639us azmon=True inv=True
- rep 3 v1: ok=True tps=5000 p99=27455us azmon=True inv=True
- rep 3 v2: ok=True tps=5000 p99=41887us azmon=True inv=True

## C5 / S1 / 5000 arrivals/s  (v1 vs v2)

pairs: 5 usable of 5 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 5 | 5,000.0 | 5,000.0 | 0.0004 | 0.0010 | +0.00% | [-0.00%, +0.00%] | CI includes 0 |
| Latency p50 (us) | 5 | 3,607.0 | 2,281.0 | 4.5 | 9.0 | -34.02% | [-37.48%, -29.83%] | v2 lower (better) |
| Latency p95 (us) | 5 | 19,471.0 | 17,887.0 | 2.3 | 65.9 | +13.55% | [-11.57%, +80.43%] | CI includes 0 |
| Latency p99 (us) | 5 | 24,575.0 | 23,599.0 | 8.7 | 84.9 | +21.62% | [-11.25%, +118.62%] | CI includes 0 |
| Latency p99.9 (us) | 5 | 56,863.0 | 29,791.0 | 100.9 | 114.0 | -65.22% | [-89.09%, +10.81%] | CI includes 0 |
| Error rate | 5 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 5 | 1,555.6 | 1,541.6 | 4.5 | 18.5 | - | - |  |
| Server write IOPS | 5 | 2,168.4 | 3,236.2 | 1.5 | 8.7 | - | - |  |
| Buffer pool hit ratio | 5 | 0.9790 | 0.9789 | 0.1045 | 0.1487 | - | - |  |
| Service time p99 (us, excl. queue) | 5 | 24,127.0 | 23,455.0 | 4.7 | 82.3 | +24.17% | [-7.32%, +116.33%] | CI includes 0 |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 5 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 v1: ok=True tps=5000 p99=23327us azmon=True inv=True
- rep 1 v2: ok=True tps=5000 p99=21455us azmon=True inv=True
- rep 2 v1: ok=True tps=5000 p99=25263us azmon=True inv=True
- rep 2 v2: ok=True tps=5000 p99=22191us azmon=True inv=True
- rep 3 v1: ok=True tps=5000 p99=24159us azmon=True inv=True
- rep 3 v2: ok=True tps=5000 p99=94079us azmon=True inv=True
- rep 4 v1: ok=True tps=5000 p99=24575us azmon=True inv=True
- rep 4 v2: ok=True tps=5000 p99=23599us azmon=True inv=True
- rep 5 v1: ok=True tps=5000 p99=28975us azmon=True inv=True
- rep 5 v2: ok=True tps=5000 p99=25519us azmon=True inv=True

## C5 / S2 / 3000 arrivals/s  (v1 vs v2)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 3,000.0 | 3,000.0 | 0.0004 | 0.0006 | -0.00% | [-0.00%, +0.00%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 14,999.0 | 12,471.0 | 2.3 | 0.3295 | -17.89% | [-19.97%, -16.69%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 22,495.0 | 17,743.0 | 4.8 | 3.2 | -22.50% | [-28.58%, -17.36%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 54,143.0 | 23,023.0 | 14.5 | 6.5 | -55.21% | [-62.46%, -43.70%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 115,519.0 | 42,143.0 | 12.3 | 2.9 | -61.70% | [-65.58%, -54.43%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 810.0 | 814.7 | 0.6274 | 0.6849 | - | - |  |
| Server write IOPS | 3 | 2,842.9 | 11,929.5 | 2.7 | 4.7 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9889 | 0.9887 | 0.0113 | 0.0077 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 52,863.0 | 22,815.0 | 13.7 | 6.5 | -54.71% | [-61.86%, -43.56%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 v1: ok=True tps=3000 p99=55999us azmon=True inv=True
- rep 1 v2: ok=True tps=3000 p99=21023us azmon=True inv=True
- rep 2 v1: ok=True tps=3000 p99=42399us azmon=True inv=True
- rep 2 v2: ok=True tps=3000 p99=23871us azmon=True inv=True
- rep 3 v1: ok=True tps=3000 p99=54143us azmon=True inv=True
- rep 3 v2: ok=True tps=3000 p99=23023us azmon=True inv=True

## C5 / S3 / 2000 arrivals/s  (v1 vs v2)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 2,614.1 | 2,693.6 | 0.4394 | 0.2069 | +3.21% | [+2.80%, +3.95%] | v2 higher (better); n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 17,295.0 | 12,543.0 | 2.3 | 0.8372 | -27.04% | [-28.97%, -24.62%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 4,284,415.0 | 3,919,871.0 | 1.1 | 0.4072 | -8.78% | [-10.19%, -7.89%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 4,390,911.0 | 4,032,511.0 | 1.1 | 0.2329 | -8.57% | [-9.95%, -7.73%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 4,481,023.0 | 4,126,719.0 | 1.3 | 1.2 | -7.96% | [-10.15%, -5.76%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 1,641.1 | 1,707.1 | 0.3619 | 0.7207 | - | - |  |
| Server write IOPS | 3 | 1,591.3 | 2,461.3 | 0.6076 | 1.9 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9745 | 0.9742 | 0.0206 | 0.0162 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 179,071.0 | 113,215.0 | 0.9036 | 0.4595 | -36.75% | [-37.49%, -36.31%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 3 | 8.0 | 7.0 | 7.5 | 8.7 | -13.10% | [-25.00%, +0.00%] | CI includes 0; n=3<5 (below protocol baseline) |
| Dropped arrivals (queue overflow) | 3 | 0.0664 | 0.0380 | 6.1 | 5.3 | -44.28% | [-50.78%, -40.59%] | v2 lower (better); n=3<5 (below protocol baseline) |

Repetition gate summary:
- rep 1 v1: ok=True tps=2598 p99=4464639us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G4_queue_delay', 'G5_errors'])
- rep 1 v2: ok=True tps=2701 p99=4020223us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G4_queue_delay', 'G5_errors'])
- rep 2 v1: ok=True tps=2620 p99=4370431us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G4_queue_delay', 'G5_errors'])
- rep 2 v2: ok=True tps=2694 p99=4032511us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G4_queue_delay', 'G5_errors'])
- rep 3 v1: ok=True tps=2614 p99=4390911us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G4_queue_delay', 'G5_errors'])
- rep 3 v2: ok=True tps=2690 p99=4038655us azmon=True inv=True FAILED: ['G4_queue_delay', 'G5_errors'] (soft-gate warnings: ['G4_queue_delay', 'G5_errors'])

## C7 / S1 / 5000 arrivals/s  (v1 vs v2)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 5,000.0 | 5,000.0 | 0.0002 | 0.0010 | +0.00% | [-0.00%, +0.00%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 3,615.0 | 2,399.0 | 0.8978 | 0.6359 | -33.67% | [-34.67%, -33.08%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 19,919.0 | 23,375.0 | 0.7729 | 1.6 | +17.66% | [+16.87%, +18.77%] | v2 higher (worse); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 24,783.0 | 32,623.0 | 3.0 | 0.7141 | +30.04% | [+26.60%, +32.75%] | v2 higher (worse); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 56,575.0 | 43,679.0 | 23.3 | 2.0 | -29.89% | [-46.05%, -19.74%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 1,638.2 | 1,647.1 | 0.2767 | 0.2118 | - | - |  |
| Server write IOPS | 3 | 2,270.8 | 3,201.1 | 0.5284 | 0.0851 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9781 | 0.9778 | 0.0119 | 0.0081 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 24,591.0 | 32,479.0 | 2.4 | 0.7152 | +30.89% | [+28.28%, +33.20%] | v2 higher (worse); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 v1: ok=True tps=5000 p99=24575us azmon=True inv=True
- rep 1 v2: ok=True tps=5000 p99=32623us azmon=True inv=True
- rep 2 v1: ok=True tps=5000 p99=24783us azmon=True inv=True
- rep 2 v2: ok=True tps=5000 p99=32431us azmon=True inv=True
- rep 3 v1: ok=True tps=5000 p99=25983us azmon=True inv=True
- rep 3 v2: ok=True tps=5000 p99=32895us azmon=True inv=True

## C7 / S4 / 3000 arrivals/s  (v1 vs v2)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 3,000.0 | 3,000.0 | 0.0003 | 0.0003 | +0.00% | [+0.00%, +0.00%] | v2 higher (better); n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 14,951.0 | 12,055.0 | 1.2 | 0.1149 | -19.17% | [-19.98%, -18.14%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 22,111.0 | 16,127.0 | 2.2 | 0.2480 | -26.43% | [-27.82%, -24.36%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 28,591.0 | 18,559.0 | 5.1 | 0.4796 | -33.36% | [-35.89%, -28.89%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 88,255.0 | 28,447.0 | 96.0 | 1.5 | -79.10% | [-92.53%, -61.11%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 520.6 | 523.8 | 1.7 | 2.4 | - | - |  |
| Server write IOPS | 3 | 2,295.4 | 14,425.8 | 1.6 | 0.9426 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9960 | 0.9958 | 0.0067 | 0.0083 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 27,567.0 | 18,399.0 | 4.4 | 0.4374 | -32.53% | [-35.37%, -28.82%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 v1: ok=True tps=3000 p99=28847us azmon=True inv=True FAILED: ['G2_storage_used'] (soft-gate warnings: ['G2_storage_used'])
- rep 1 v2: ok=True tps=3000 p99=18495us azmon=True inv=True FAILED: ['G2_storage_used'] (soft-gate warnings: ['G2_storage_used'])
- rep 2 v1: ok=True tps=3000 p99=28591us azmon=True inv=True FAILED: ['G2_storage_used'] (soft-gate warnings: ['G2_storage_used'])
- rep 2 v2: ok=True tps=3000 p99=18559us azmon=True inv=True FAILED: ['G2_storage_used'] (soft-gate warnings: ['G2_storage_used'])
- rep 3 v1: ok=True tps=3000 p99=26255us azmon=True inv=True FAILED: ['G2_storage_used'] (soft-gate warnings: ['G2_storage_used'])
- rep 3 v2: ok=True tps=3000 p99=18671us azmon=True inv=True FAILED: ['G2_storage_used'] (soft-gate warnings: ['G2_storage_used'])
