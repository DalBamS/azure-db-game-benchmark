# azure-db-game-benchmark

게임 OLTP 워크로드로 Azure Database를 비교한 벤치마크 실험실.

- **실험 B**: Azure Database for MySQL Flexible Server — Premium SSD **v1 vs v2** (same-IOPS 5,000)
- **실험 A**: **PostgreSQL Flexible Server vs Azure HorizonDB** (동일 vCore·동일 부하)

모든 보고 수치는 10가지 유효성 게이트(G1~G10)를 통과한 측정에서만 산출한다.
설계 원칙과 게이트 정의는 [benchmark-prompt-v2.md](benchmark-prompt-v2.md), 전체 실행 이력은 [docs/run-log.md](docs/run-log.md) 참고.

## 저장소 구성

| 경로 | 내용 |
|---|---|
| `cmd/gamebench`, `internal/bench` | 부하 도구 (Go). MySQL/PostgreSQL 공용, open-loop, HDR 히스토그램, 게이트 자동 판정, `-ro-dsn` 복제본 라우팅, `-checkpoint` 장기 실행 스냅샷 |
| `infra/expA/main.bicep` | 실험 A 전체 스택 (PG + HorizonDB + 벤치 VM + 네트워크) — 검증 완료 |
| `infra/expB/main.bicep` | 실험 B 전체 스택 (MySQL v1/v2 쌍 + 벤치 VM) — 검증 완료 |
| `infra/monitoring/` | 모니터링 VM cloud-init (Telegraf+InfluxDB+Grafana), 대시보드 JSON 2종 |
| `scripts/vm*/` | VM 측 오케스트레이션 (적재 / knee 탐색 / 본 측정 / soak) |
| `scripts/*.py` | 분석 파이프라인: `azmon_slice.py`(서버 지표 절단), `analyze.py`(paired 효과+bootstrap CI), `charts.py`, `build_pptx*.py`, `load_history_to_influx.py` |
| `results/` | 측정 원본 JSON (클라이언트 지표 + 서버 status 시계열 + Azure Monitor 슬라이스 + 게이트 판정) |
| `analysis/` | 통계 결과(md/json/csv)와 차트 |
| `docs/` | 실험 계획서, 실행 로그, 발표 스크립트 |
| `report/` | 최종 보고서 PPTX (KRAFTON 디자인 한글판 / 엔지니어용 영문판) |

## 재현 절차 (요약)

사전 조건: Azure CLI 로그인, 구독에 `Microsoft.DBforMySQL`/`DBforPostgreSQL`/`HorizonDB` provider 등록.
Premium SSD v2(MySQL)와 HorizonDB는 **프리뷰**라 구독 등록/지원 리전이 필요하다 (HorizonDB 리전: australiaeast 등 6개; 16 vCore 스케일은 용량 제약을 겪을 수 있음 — 8 vCore 경유가 안정적).

```bash
# 0) 빌드
GOOS=linux GOARCH=amd64 go build -o build/gamebench-linux-amd64 ./cmd/gamebench

# 1) 인프라 (RG는 미리 생성; 검증만 하려면 create 대신 validate)
az deployment group create -g <rg-expB> -f infra/expB/main.bicep \
  -p administratorLoginPassword='<강한비밀번호>' benchAdminSshKey="$(cat ~/.ssh/id.pub)" operatorIp=<내IP>
az deployment group create -g <rg-expA> -f infra/expA/main.bicep \
  -p administratorLoginPassword='<강한비밀번호>' benchAdminSshKey="$(cat ~/.ssh/id.pub)" operatorIp=<내IP>

# 2) 벤치 VM 준비 (scp로 binary + scripts/vm* 업로드, ~/.bench/*.env는 cloud-init이 생성)
scp build/gamebench-linux-amd64 benchadmin@<benchIp>:~/gamebench
scp scripts/vm/*.sh benchadmin@<benchIp>:~/scripts/     # 실험 B (실험 A는 scripts/vm-pg)

# 3) 데이터 적재 → knee 탐색 → 본 측정 (VM에서)
bash ~/scripts/load-both.sh 2500000 20 1024 320          # ~24GiB (G1: 캐시 8GiB의 3배)
bash ~/scripts/knee-both.sh C1 S1                        # 한계점 → 권장 도착률(65%) 출력
bash ~/scripts/run-case.sh C1 S1 5000 5 600 256 120      # 도착률 5000/s, 5회 × 10분, 256 conn

# 4) 분석 (로컬)
python scripts/azmon_slice.py "results/**/rep*.json"     # 서버 지표를 UTC 창으로 절단(G6)
python scripts/analyze.py results/expB --out analysis/expB
python scripts/charts.py results/expB analysis/expB/charts
python scripts/build_pptx_kr.py report/보고서.pptx
```

## 운영 시 주의 (실측 기반)

- **HA를 켜면 storage autoGrow가 강제**된다. 지속 쓰기 부하에서 스토리지가 자동 증설되며(실측: 14h soak에 128GiB→1.6TB), **MySQL은 축소가 불가능**하다. 장기 부하 전 반드시 인지할 것.
- v6 컴퓨팅과 SSD v2는 플랫폼이 상호 강제한다(`UnsupportedStorageSkuForV6VmSku`). 컴퓨팅 세대 confound는 제거 불가.
- SSD v2 서버는 읽기 복제본 생성이 현재 InternalServerError로 실패한다(프리뷰 제약).
- 128GiB에서 v1의 무료 IOPS ≈ 5,037 → 이 크기에선 as-provisioned = same-IOPS.
- MCAPS류 governance가 야간에 VM/DB를 자동 종료할 수 있다. 장기 soak는 `-checkpoint`(30분 스냅샷)와 함께 돌릴 것.
- 시간대는 전부 UTC로 기록·조회한다 (이전 실험의 −9h 버그 재발 방지).

## 결과 하이라이트

`analysis/*/analysis.md`와 `report/` 참고. 요약: MySQL은 v2가 평균 −30%·쓰기 p99 −55%·장시간 꼬리 우세, 단 No-HA 단기 꼬리는 열세. HorizonDB는 동일 vCore에서 p99 −75~−95%지만 캐시 5배·1.8배 요금·프리뷰 조건부. PG는 P15 디스크에서 쓰기·대용량 부하 시 스톨 실측 — 디스크 상향이 선결.
