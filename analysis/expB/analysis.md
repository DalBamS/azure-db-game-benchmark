# v1 vs v2 paired analysis

Generated 2026-08-17T20:00:58.811049+00:00

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

Repetition gate summary:
- rep 1 v1: ok=True tps=5000 p99=23167us azmon=True inv=True
- rep 1 v2: ok=True tps=5000 p99=17759us azmon=True inv=True
- rep 2 v1: ok=True tps=5000 p99=33471us azmon=True inv=True
- rep 2 v2: ok=True tps=5000 p99=18511us azmon=True inv=True
- rep 3 v1: ok=True tps=5000 p99=25199us azmon=True inv=True
- rep 3 v2: ok=True tps=5000 p99=64607us azmon=True inv=True

## C5 / S1 / 5000 arrivals/s  (v1 vs v2)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | v1 median | v2 median | v1 CV% | v2 CV% | Δ% (v2 vs v1, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 5,000.0 | 5,000.0 | 0.0003 | 0.0013 | +0.00% | [-0.00%, +0.00%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 3,743.0 | 2,281.0 | 4.8 | 11.5 | -33.62% | [-39.89%, -26.45%] | v2 lower (better); n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 19,167.0 | 17,119.0 | 0.9915 | 73.8 | +29.82% | [-12.64%, +184.82%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 24,159.0 | 22,191.0 | 4.0 | 90.9 | +46.53% | [-12.16%, +289.42%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 56,735.0 | 28,335.0 | 124.7 | 112.2 | -52.25% | [-93.90%, +221.94%] | CI includes 0; n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 1,657.9 | 1,538.5 | 5.0 | 23.1 | - | - |  |
| Server write IOPS | 3 | 2,138.4 | 3,242.2 | 2.0 | 11.7 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9778 | 0.9789 | 0.1135 | 0.1969 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 23,935.0 | 22,047.0 | 2.2 | 88.7 | +46.85% | [-8.62%, +276.21%] | CI includes 0; n=3<5 (below protocol baseline) |

Repetition gate summary:
- rep 1 v1: ok=True tps=5000 p99=23327us azmon=True inv=True
- rep 1 v2: ok=True tps=5000 p99=21455us azmon=True inv=True
- rep 2 v1: ok=True tps=5000 p99=25263us azmon=True inv=True
- rep 2 v2: ok=True tps=5000 p99=22191us azmon=True inv=True
- rep 3 v1: ok=True tps=5000 p99=24159us azmon=True inv=True
- rep 3 v2: ok=True tps=5000 p99=94079us azmon=True inv=True
