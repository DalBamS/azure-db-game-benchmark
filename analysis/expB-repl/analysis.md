# v1 vs v2 paired analysis

Generated 2026-08-25T22:10:36.072546+00:00

## C4 / S1 / 5000 arrivals/s  (routed vs primary)

pairs: 3 usable of 3 (hard-gate-failed pairs excluded; soft gates: none)

| Outcome | n | routed median | primary median | routed CV% | primary CV% | Δ% (primary vs routed, geo-mean) | 95% CI | note |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Throughput (success TPS) | 3 | 5,000.0 | 5,000.0 | 0.0001 | 0.0008 | -0.00% | [-0.00%, +0.00%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p50 (us) | 3 | 3,823.0 | 3,991.0 | 4.3 | 1.2 | +1.35% | [-2.92%, +4.60%] | CI includes 0; n=3<5 (below protocol baseline) |
| Latency p95 (us) | 3 | 22,911.0 | 23,599.0 | 0.7570 | 1.1 | +2.89% | [+1.04%, +4.09%] | primary higher (worse); n=3<5 (below protocol baseline) |
| Latency p99 (us) | 3 | 29,007.0 | 30,399.0 | 2.4 | 7.2 | +8.03% | [+4.80%, +14.69%] | primary higher (worse); n=3<5 (below protocol baseline) |
| Latency p99.9 (us) | 3 | 89,279.0 | 204,287.0 | 55.6 | 42.4 | +123.51% | [+31.36%, +234.34%] | primary higher (worse); n=3<5 (below protocol baseline) |
| Error rate | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |
| Server read IOPS | 3 | 603.3 | 1,775.9 | 0.9497 | 2.1 | - | - |  |
| Server write IOPS | 3 | 2,232.3 | 2,217.8 | 0.8351 | 0.5427 | - | - |  |
| Buffer pool hit ratio | 3 | 0.9863 | 0.9761 | 0.0143 | 0.0422 | - | - |  |
| Service time p99 (us, excl. queue) | 3 | 28,511.0 | 29,311.0 | 1.6 | 5.3 | +5.78% | [+2.81%, +10.82%] | primary higher (worse); n=3<5 (below protocol baseline) |
| Burst recovery time (s, S3 only) | 0 | - | - | - | - | - | - | insufficient pairs for effect/CI |
| Dropped arrivals (queue overflow) | 3 | 0.0000 | 0.0000 | - | - | - | - | insufficient pairs for effect/CI |

Repetition gate summary:
- rep 1 routed: ok=True tps=5000 p99=29007us azmon=True inv=True
- rep 1 primary: ok=True tps=5000 p99=30399us azmon=True inv=True
- rep 2 routed: ok=True tps=5000 p99=28079us azmon=True inv=True
- rep 2 primary: ok=True tps=5000 p99=29455us azmon=True inv=True
- rep 3 routed: ok=True tps=5000 p99=29407us azmon=True inv=True
- rep 3 primary: ok=True tps=5000 p99=33727us azmon=True inv=True
