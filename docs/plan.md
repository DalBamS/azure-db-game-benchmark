# 실험 계획서 — 실험 B: Azure MySQL Flexible Server Premium SSD v1 vs v2 (Phase 0: C3)

작성: 2026-08-18 · 상태: **실행 전 검토용 (v1)** · 근거 프롬프트: `../benchmark-prompt-v2.md`

## 1. 질문

게임 OLTP 워크로드에서, 동일 compute(8 vCore / 64 GiB, MemoryOptimized)·동일 데이터·동일 부하 조건일 때
**Premium SSD v2 스토리지가 v1 대비 처리량·지연·안정성에서 얼마나 다른가** (효과 크기 + 95% CI).

## 2. 변인

| 구분 | 항목 | 값 / 처리 |
|---|---|---|
| **조작 변인** | 스토리지 SKU | v1 `Premium_LRS` vs v2 `PremiumV2_LRS` |
| **통제 변인** | compute 등급 | 8 vCore / 64 GiB, MemoryOptimized (E8ds) |
| | MySQL 버전 | 8.4.9 양쪽 동일 |
| | `innodb_buffer_pool_size` | **양쪽 16 GiB(17179869184)로 고정** — 기본값 48 GiB를 유지하면 G1(데이터 ≥ 버퍼풀 2배)을 위해 ≥96 GiB 적재가 필요해 시간·비용상 비현실적. 이 축소는 사전등록된 통제이며 보고서 confound 표에 명시 |
| | `innodb_redo_log_capacity` | 양쪽 동일 값(현재값 확인 후 고정) |
| | 스토리지 크기 | 양쪽 128 GiB |
| | Provisioned IOPS (**셀 (b) same-IOPS**) | 양쪽 5,000 IOPS (v1은 추가 IOPS 구매, v2는 5,000/125 MBps) — **Phase 0 주 셀** |
| | Provisioned IOPS (**셀 (a) as-provisioned**) | v1 = 128 GiB 기본 free IOPS, v2 = 5,000 — Phase 1에서 실행 |
| | HA / zone / replica | Disabled / zone 1 / 없음 (C3) |
| | 네트워크 경로 | 양쪽 public endpoint + TLS, 벤치 VM NAT IP 허용, 동일 리전(koreacentral) |
| | 데이터셋 | 동일 시드·동일 크기, 랜덤 바이트 payload |
| | 부하 | 동일 도착률(open-loop), 동일 시나리오, 교대 실행 |
| **통제 불가 confound (기록)** | compute 세대 | v1 = `Standard_E8ds_v5`, v2 = `Standard_E8ds_v6` (플랫폼이 v2 스토리지에 v6 강제) |
| | `innodb_doublewrite` | 플랫폼 read-only: v1 OFF / v2 ON 관측(이전 실험) → 실행 전 재확인해 기록 |
| | `innodb_flush_log_at_trx_commit`, `sync_binlog` | read-only, 값 기록 |
| **종속 변인** | 처리량(성공 TPS), p50/p95/p99/p99.9 지연(µs, HDR), 오류율/타임아웃율, 풀 대기, 서버 read/write IOPS, 버퍼풀 히트율, CPU%, io_consumption_percent, 지연 시계열 | |

## 3. 데이터셋 (G1)

- 목표 실제 크기 ≥ 32 GiB (버퍼풀 16 GiB × 2), 설계 목표 ~40 GiB.
- 스키마: `accounts`(profile 1,024 B 랜덤) · `inventory`(계정당 20슬롯, attrs 256 B 랜덤) · `game_sessions` · `purchase_ledger` · `match_results` · `leaderboard` · `guilds` · `guild_members`.
- 규모 파라미터: accounts = 5,000,000 → accounts ≈ 5.5 GiB, inventory 1억 행 ≈ 30 GiB(+인덱스). 적재 후 `information_schema` 실측값을 결과에 기록하고 G1을 자동 판정.
- payload는 `crypto/rand` 기반 랜덤 바이트(압축 불가).
- 키 분포: Zipf(θ 조정), 상위 1키 ≤ 5%(S1~S3), S4 핫스팟만 별도 계수. 실제 상위 1/100/1,000키 점유율을 로그로 기록.

## 4. 시나리오와 오퍼레이션 믹스

| 시나리오 | 믹스 | 목적 |
|---|---|---|
| S1 일반 플레이 | profile_read 30 / inventory_read 20 / leaderboard_read 10 / guild_read 5 / session_upsert 10 / inventory_update 12 / match_result 8 / purchase 5 (읽기 65 / 쓰기 35) | 주 결과 |
| S2 쓰기 집중 | match_result 30 / inventory_update 25 / purchase 15 / session_upsert 10 / profile_read 15 / leaderboard_read 5 (쓰기 80) | 스토리지 쓰기 경로 |
| S3 로그인 피크 | S1 믹스로 정상→4배 버스트(3분)→복귀, 복구 시간 측정 | 버스트 흡수 |
| S4 핫스팟 | S2 믹스, Zipf 계수 상향(상위 1키 ~10%) | 잠금 경합 하의 스토리지 |

Phase 0 완료 기준은 **S1**. S2~S4는 Phase 0 파이프라인 통과 후 순차 추가.

## 5. 부하 모델 및 실행 절차 (케이스당)

1. **Knee 탐색** (closed-loop): 동시성 16→32→64→128→256→512, 단계당 2분, think 0. 처리량 증가 <10% & p99 2배 이상이면 knee. 결과 표에 미포함.
2. **본 측정** (open-loop): 도착률 = knee 처리량 × 0.65. warmup → G7 steady-state 확인 → **10분 측정**. arm 교대(v1→v2, v2→v1, …) **5회 반복**, run 간 cooldown 2분.
3. **스트레스**: knee × 1.0, 1.2, 각 10분 × 3회. G5 초과 시 "지속 불가"로 표기.
4. 커넥션 풀 = 최대 in-flight = 워커 수(기본 512), 풀 대기 p99 측정(G4). 도착 큐는 10초 분량으로 제한, 초과 drop은 게이트로 기록.
5. 지연 = 예정 도착 시각부터 완료까지(HDR, µs). 오류·타임아웃은 별도 집계, 지연 분포에서 제외.
6. 벤치 도구가 5초 간격으로 `SHOW GLOBAL STATUS`(Innodb_data_reads/writes, buffer pool reads/read_requests, Threads_running 등)를 대상 서버에서 직접 샘플링해 결과 JSON에 포함 → G2/G6를 외부 모니터링에 의존하지 않고 판정. Azure Monitor 플랫폼 지표(cpu_percent, io_consumption_percent, storage_io_count, memory_percent, active_connections)는 run UTC 창으로 잘라 별도 첨부.

## 6. 유효성 게이트 (프롬프트 §3 그대로, 자동 판정)

G1 데이터 ≥ 버퍼풀 2배 · G2 read IOPS > 0 & 버퍼풀 히트율 < 99% · G3 벤치 VM CPU < 60% & 네트워크 < 50% · G4 풀 대기 p99 < 1 ms · G5 오류 < 1%, 타임아웃 < 0.5% · G6 서버 지표 창 ±1분 & 비어 있지 않음(UTC) · G7 warmup 후 2분 CV < 5% · G8 arm 간 warm 대칭 · G9 앱 수준 불변식(잔액 = 초기 − 구매 합, 인벤토리 수량 ≥ 0 은 애플리케이션 검증) · G10 실행 시간 ≥ 95%.

## 7. 통계

셀 = (케이스, 시나리오, IOPS 셀, 부하 수준). 반복 i의 (v1_i, v2_i) paired `log(v2/v1)` → 기하평균 % 변화, bootstrap(10,000, seed 1234) 95% percentile CI, 중앙값/MAD, 반복 간 CV. 지연은 p50/p95/p99/p99.9 각각. n<5 셀은 "미완"으로 표기.

## 8. 인프라 (기존 재사용 + 수정)

| 리소스 | 조치 |
|---|---|
| `mysqlbm-euson-v1` (E8ds_v5, Premium_LRS 64 GiB, 492 IOPS) | 스토리지 128 GiB, IOPS 5,000, buffer pool 16 GiB |
| `mysqlbm-euson-v2` (E8ds_v6, PremiumV2_LRS 64 GiB, 5,000 IOPS) | 스토리지 128 GiB, buffer pool 16 GiB |
| `mysqlbm-euson-lg-vm` (D4ds_v5, 정지) | D8ds_v5로 리사이즈, 벤치 VM |
| 모니터링 VM (신규, D4s_v5, 같은 VNet) | Telegraf(inputs.mysql ×2, bench VM 지표) → InfluxDB 2.x → Grafana(대시보드 provisioning) |
| Managed Grafana / ADX(이전 실험) | 미사용. 정리 여부는 사용자 결정 |

## 9. 예산·시간

- 상한: USD 2,000 / 7일 (사용자 지정). Phase 0 예상: DB 2대(~USD 1.3/h ×2) + VM 2대(~USD 0.5/h) ≈ USD 3/h → 24h ≈ USD 75, IOPS 추가·스토리지 소액.
- Phase 0 예상 소요: 서버 조정 30분 + 데이터 적재 1–2h + knee 1h + S1 본 측정 2.5h + 게이트 반복 여유 → 약 8h.
- 전체 8케이스 S1+S2 확장 시 ~3일(순차) — 시간 초과 시 케이스를 줄이되 게이트 통과 run만 보고.

## 10. 산출물

`results/<run-id>/` — run.json(클라이언트 지표 + 서버 status 시계열 + Azure Monitor slice + 게이트 판정 + 환경 스냅샷), `docs/run-log.md`, `analysis/` (CSV/JSON/MD), 최종 PPTX.
