package bench

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

// ---- PostgreSQL / HorizonDB dialect -------------------------------------------

var PGSchema = []string{
	`CREATE TABLE IF NOT EXISTS accounts (
		id BIGINT PRIMARY KEY, username VARCHAR(32) NOT NULL, level INT NOT NULL, exp BIGINT NOT NULL,
		balance BIGINT NOT NULL, guild_id INT NOT NULL, profile BYTEA NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp())`,
	`CREATE INDEX IF NOT EXISTS ix_accounts_guild ON accounts(guild_id)`,
	`CREATE TABLE IF NOT EXISTS inventory (
		account_id BIGINT NOT NULL, slot SMALLINT NOT NULL, item_id INT NOT NULL, qty INT NOT NULL,
		version INT NOT NULL DEFAULT 0, attrs BYTEA NOT NULL, PRIMARY KEY (account_id, slot))`,
	`CREATE TABLE IF NOT EXISTS game_sessions (
		account_id BIGINT PRIMARY KEY, session_token BYTEA NOT NULL, client_version VARCHAR(16) NOT NULL,
		last_seen TIMESTAMPTZ NOT NULL, login_count INT NOT NULL DEFAULT 1)`,
	`CREATE TABLE IF NOT EXISTS purchase_ledger (
		request_id BYTEA PRIMARY KEY, account_id BIGINT NOT NULL, item_id INT NOT NULL, amount BIGINT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp())`,
	`CREATE INDEX IF NOT EXISTS ix_ledger_account ON purchase_ledger(account_id, created_at)`,
	`CREATE TABLE IF NOT EXISTS match_results (
		match_id BIGSERIAL PRIMARY KEY, account_id BIGINT NOT NULL, score INT NOT NULL, duration_ms INT NOT NULL,
		payload BYTEA NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp())`,
	`CREATE INDEX IF NOT EXISTS ix_match_account ON match_results(account_id, created_at)`,
	`CREATE TABLE IF NOT EXISTS leaderboard (
		account_id BIGINT PRIMARY KEY, score BIGINT NOT NULL, matches INT NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp())`,
	`CREATE INDEX IF NOT EXISTS ix_lb_score ON leaderboard(score DESC)`,
	`CREATE TABLE IF NOT EXISTS guilds (guild_id INT PRIMARY KEY, name VARCHAR(64) NOT NULL, rating INT NOT NULL, notice BYTEA NOT NULL)`,
	`CREATE TABLE IF NOT EXISTS guild_members (guild_id INT NOT NULL, account_id BIGINT NOT NULL, role SMALLINT NOT NULL, contribution INT NOT NULL, PRIMARY KEY (guild_id, account_id))`,
	`CREATE TABLE IF NOT EXISTS bench_meta (k VARCHAR(64) PRIMARY KEY, v VARCHAR(255) NOT NULL)`,
}

func pgOpProfileRead() Op {
	return Op{Name: "profile_read", Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		var level int
		var exp, bal int64
		var profile []byte
		return c.QueryRowContext(ctx, `SELECT level, exp, balance, profile FROM accounts WHERE id=$1`, w.Keys.Pick()).Scan(&level, &exp, &bal, &profile)
	}}
}
func pgOpInventoryRead() Op {
	return Op{Name: "inventory_read", Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		rows, err := c.QueryContext(ctx, `SELECT slot, item_id, qty, attrs FROM inventory WHERE account_id=$1`, w.Keys.Pick())
		if err != nil {
			return err
		}
		defer rows.Close()
		var slot, item, qty int
		var attrs []byte
		for rows.Next() {
			if err := rows.Scan(&slot, &item, &qty, &attrs); err != nil {
				return err
			}
		}
		return rows.Err()
	}}
}
func pgOpLeaderboardRead() Op {
	return Op{Name: "leaderboard_read", Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		rows, err := c.QueryContext(ctx, `SELECT account_id, score FROM leaderboard ORDER BY score DESC LIMIT 100`)
		if err != nil {
			return err
		}
		defer rows.Close()
		var id, score int64
		for rows.Next() {
			if err := rows.Scan(&id, &score); err != nil {
				return err
			}
		}
		return rows.Err()
	}}
}
func pgOpGuildRead() Op {
	return Op{Name: "guild_read", Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		rows, err := c.QueryContext(ctx, `SELECT g.guild_id, g.name, g.rating, gm.account_id, gm.role, gm.contribution
			FROM accounts a JOIN guilds g ON g.guild_id=a.guild_id JOIN guild_members gm ON gm.guild_id=g.guild_id
			WHERE a.id=$1 LIMIT 50`, w.Keys.Pick())
		if err != nil {
			return err
		}
		defer rows.Close()
		var gid, rating, role, contrib int
		var name string
		var aid int64
		for rows.Next() {
			if err := rows.Scan(&gid, &name, &rating, &aid, &role, &contrib); err != nil {
				return err
			}
		}
		return rows.Err()
	}}
}
func pgOpSessionUpsert() Op {
	return Op{Name: "session_upsert", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		_, err := c.ExecContext(ctx, `INSERT INTO game_sessions(account_id, session_token, client_version, last_seen)
			VALUES ($1, $2, '1.0.0', clock_timestamp()) ON CONFLICT (account_id) DO UPDATE SET last_seen=EXCLUDED.last_seen, login_count=game_sessions.login_count+1`,
			w.Keys.Pick(), uuidBytes(w.Rng))
		return err
	}}
}
func pgOpInventoryUpdate() Op {
	return Op{Name: "inventory_update", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		slot := 1 + w.Rng.IntN(w.Slots)
		tx, err := c.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var qty int
		if err := tx.QueryRowContext(ctx, `SELECT qty FROM inventory WHERE account_id=$1 AND slot=$2 FOR UPDATE`, id, slot).Scan(&qty); err != nil {
			return err
		}
		delta := 1 + w.Rng.IntN(5)
		if qty-delta < 0 {
			delta = 0
		}
		if _, err := tx.ExecContext(ctx, `UPDATE inventory SET qty=qty-$1, version=version+1, attrs=$2 WHERE account_id=$3 AND slot=$4`,
			delta, w.randBytes(256), id, slot); err != nil {
			return err
		}
		return tx.Commit()
	}}
}
func pgOpMatchResult() Op {
	return Op{Name: "match_result", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		score := w.Rng.IntN(1000)
		tx, err := c.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if _, err := tx.ExecContext(ctx, `INSERT INTO match_results(account_id, score, duration_ms, payload) VALUES ($1,$2,$3,$4)`,
			id, score, 30000+w.Rng.IntN(600000), w.randBytes(128)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE leaderboard SET score=score+$1, matches=matches+1, updated_at=clock_timestamp() WHERE account_id=$2`, score, id); err != nil {
			return err
		}
		return tx.Commit()
	}}
}
func pgOpPurchase() Op {
	return Op{Name: "purchase", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		item := 1 + w.Rng.IntN(5000)
		amount := int64(10 + w.Rng.IntN(90))
		tx, err := c.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		res, err := tx.ExecContext(ctx, `UPDATE accounts SET balance=balance-$1, updated_at=clock_timestamp() WHERE id=$2 AND balance>=$1`, amount, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return tx.Commit()
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO purchase_ledger(request_id, account_id, item_id, amount) VALUES ($1,$2,$3,$4)`,
			uuidBytes(w.Rng), id, item, amount); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE inventory SET qty=qty+1, version=version+1 WHERE account_id=$1 AND slot=$2`, id, 1+w.Rng.IntN(w.Slots)); err != nil {
			return err
		}
		return tx.Commit()
	}}
}
func pgOpLogin() Op {
	pr, su, ir := pgOpProfileRead(), pgOpSessionUpsert(), pgOpInventoryRead()
	return Op{Name: "login", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		if err := pr.Run(ctx, c, w); err != nil {
			return err
		}
		if err := su.Run(ctx, c, w); err != nil {
			return err
		}
		return ir.Run(ctx, c, w)
	}}
}

func PGScenarioMix(name string) (*Mix, error) {
	switch name {
	case "S1", "s1", "normal":
		return &Mix{Ops: []Op{pgOpProfileRead(), pgOpInventoryRead(), pgOpLeaderboardRead(), pgOpGuildRead(), pgOpSessionUpsert(), pgOpInventoryUpdate(), pgOpMatchResult(), pgOpPurchase()},
			Weights: []int{30, 20, 10, 5, 10, 12, 8, 5}}, nil
	case "S2", "s2", "write", "S4", "s4", "hotspot":
		return &Mix{Ops: []Op{pgOpMatchResult(), pgOpInventoryUpdate(), pgOpPurchase(), pgOpSessionUpsert(), pgOpProfileRead(), pgOpLeaderboardRead()},
			Weights: []int{30, 25, 15, 10, 15, 5}}, nil
	case "S3", "s3", "login":
		return &Mix{Ops: []Op{pgOpLogin(), pgOpProfileRead(), pgOpInventoryRead(), pgOpInventoryUpdate(), pgOpMatchResult()}, Weights: []int{40, 20, 15, 15, 10}}, nil
	case "read":
		return &Mix{Ops: []Op{pgOpProfileRead(), pgOpInventoryRead()}, Weights: []int{50, 50}}, nil
	}
	return nil, fmt.Errorf("unknown scenario %q", name)
}

// ---- PG status / env --------------------------------------------------------

// PGFetchStatus maps PostgreSQL counters onto the same keys the MySQL sampler uses so
// ComputeDelta/gates work unchanged: Innodb_data_reads <- blks_read (physical reads),
// buffer_pool_read_requests <- blks_hit+blks_read, data_writes <- buffers written
// (checkpoint+clean+backend), Com_commit <- xact_commit, Threads_running <- active backends.
func PGFetchStatus(ctx context.Context, db *sql.DB) (map[string]uint64, error) {
	out := map[string]uint64{}
	var blksRead, blksHit, xactCommit, xactRollback, tupReturned, tupInserted, tupUpdated, tempBytes int64
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(blks_read),0), COALESCE(SUM(blks_hit),0), COALESCE(SUM(xact_commit),0), COALESCE(SUM(xact_rollback),0),
		COALESCE(SUM(tup_returned),0), COALESCE(SUM(tup_inserted),0), COALESCE(SUM(tup_updated),0), COALESCE(SUM(temp_bytes),0) FROM pg_stat_database WHERE datname=current_database()`).
		Scan(&blksRead, &blksHit, &xactCommit, &xactRollback, &tupReturned, &tupInserted, &tupUpdated, &tempBytes); err != nil {
		return nil, err
	}
	out["Innodb_data_reads"] = uint64(blksRead)
	out["Innodb_data_read"] = uint64(blksRead) * 8192
	out["Innodb_buffer_pool_reads"] = uint64(blksRead)
	out["Innodb_buffer_pool_read_requests"] = uint64(blksRead + blksHit)
	out["Com_commit"] = uint64(xactCommit)
	out["Com_rollback"] = uint64(xactRollback)
	out["Innodb_rows_read"] = uint64(tupReturned)
	out["Innodb_rows_inserted"] = uint64(tupInserted)
	out["Innodb_rows_updated"] = uint64(tupUpdated)
	out["Questions"] = uint64(xactCommit + xactRollback)
	// writes: PG17 moved checkpointer stats; try pg_stat_io (PG16+) first
	var writes, walBytes int64
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(writes),0) FROM pg_stat_io WHERE object='relation'`).Scan(&writes); err == nil {
		out["Innodb_data_writes"] = uint64(writes)
		out["Innodb_data_written"] = uint64(writes) * 8192
	}
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(wal_bytes,0) FROM pg_stat_wal`).Scan(&walBytes); err == nil {
		out["Innodb_os_log_written"] = uint64(walBytes)
	}
	var active, total int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER (WHERE state='active'), COUNT(*) FROM pg_stat_activity WHERE backend_type='client backend'`).Scan(&active, &total); err == nil {
		out["Threads_running"] = uint64(active)
		out["Threads_connected"] = uint64(total)
	}
	var lockWaits int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type='Lock'`).Scan(&lockWaits); err == nil {
		out["Innodb_row_lock_waits"] = uint64(lockWaits) // instantaneous, not cumulative (documented)
	}
	return out, nil
}

var pgEnvVars = []string{"server_version", "shared_buffers", "effective_cache_size", "work_mem", "max_connections", "wal_level",
	"synchronous_commit", "fsync", "full_page_writes", "checkpoint_timeout", "max_wal_size", "wal_compression", "huge_pages",
	"random_page_cost", "effective_io_concurrency", "max_worker_processes", "ssl", "azure.extensions", "block_size"}

func PGCaptureEnv(ctx context.Context, db *sql.DB, schema string) (*EnvSnapshot, error) {
	e := &EnvSnapshot{Variables: map[string]string{}, TableGiB: map[string]float64{}, Rows: map[string]int64{}, CapturedAt: time.Now().UTC()}
	for _, v := range pgEnvVars {
		var val string
		if err := db.QueryRowContext(ctx, "SELECT current_setting($1, true)", v).Scan(&val); err == nil {
			e.Variables[v] = val
		}
	}
	// express shared_buffers in bytes under the MySQL key so G1 works
	if sb, ok := e.Variables["shared_buffers"]; ok {
		if bytes, err := pgSizeToBytes(sb, e.Variables["block_size"]); err == nil {
			e.Variables["innodb_buffer_pool_size"] = strconv.FormatInt(bytes, 10)
		}
	}
	rows, err := db.QueryContext(ctx, `SELECT relname, pg_total_relation_size(c.oid)/1073741824.0, GREATEST(reltuples,0)::bigint
		FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND relkind='r'`)
	if err != nil {
		return e, err
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		var gib float64
		var n int64
		if err := rows.Scan(&t, &gib, &n); err != nil {
			return e, err
		}
		e.TableGiB[t] = gib
		e.Rows[t] = n
		e.TotalGiB += gib
	}
	return e, rows.Err()
}

func pgSizeToBytes(s, blockSize string) (int64, error) {
	s = strings.TrimSpace(s)
	units := map[string]int64{"kB": 1024, "MB": 1 << 20, "GB": 1 << 30, "TB": 1 << 40, "B": 1}
	for u, m := range units {
		if strings.HasSuffix(s, u) {
			n, err := strconv.ParseInt(strings.TrimSpace(strings.TrimSuffix(s, u)), 10, 64)
			return n * m, err
		}
	}
	// plain number = blocks
	n, err := strconv.ParseInt(s, 10, 64)
	bs, _ := strconv.ParseInt(blockSize, 10, 64)
	if bs == 0 {
		bs = 8192
	}
	return n * bs, err
}

// PGLoad mirrors Load for PostgreSQL (multi-row inserts with $n placeholders).
func PGLoad(ctx context.Context, db *sql.DB, cfg LoadConfig, log func(string, ...any)) error {
	if cfg.Workers <= 0 {
		cfg.Workers = 16
	}
	if cfg.Batch <= 0 {
		cfg.Batch = 500
	}
	start := time.Now()
	r := rand.New(rand.NewPCG(cfg.Seed, 1))
	buf := make([]byte, 1024)
	{
		var sb strings.Builder
		args := []any{}
		for g := 1; g <= cfg.Guilds; g++ {
			if sb.Len() == 0 {
				sb.WriteString("INSERT INTO guilds(guild_id,name,rating,notice) VALUES ")
			} else {
				sb.WriteString(",")
			}
			fill(buf, r)
			n := len(args)
			fmt.Fprintf(&sb, "($%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4)
			args = append(args, g, fmt.Sprintf("guild_%d", g), 1000+r.IntN(2000), append([]byte(nil), buf...))
			if g%500 == 0 || g == cfg.Guilds {
				if _, err := db.ExecContext(ctx, sb.String(), args...); err != nil {
					return fmt.Errorf("guilds: %w", err)
				}
				sb.Reset()
				args = args[:0]
			}
		}
		log("guilds loaded: %d", cfg.Guilds)
	}
	type job struct{ lo, hi int64 }
	jobs := make(chan job, cfg.Workers*2)
	errCh := make(chan error, cfg.Workers)
	var done int64
	var wg = make(chan struct{}, cfg.Workers)
	for w := 0; w < cfg.Workers; w++ {
		go func(id int) {
			defer func() { wg <- struct{}{} }()
			rr := rand.New(rand.NewPCG(cfg.Seed+uint64(id)*31337, uint64(id)))
			prof := make([]byte, cfg.ProfileBytes)
			attrs := make([]byte, cfg.AttrsBytes)
			conn, err := db.Conn(ctx)
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()
			for j := range jobs {
				if err := pgLoadRange(ctx, conn, cfg, rr, prof, attrs, j.lo, j.hi); err != nil {
					errCh <- fmt.Errorf("range %d-%d: %w", j.lo, j.hi, err)
					return
				}
				d := j.hi
				done = d
				if (j.lo/int64(cfg.Batch))%200 == 0 {
					el := time.Since(start).Seconds()
					log("loaded ~%d/%d accounts (%.0f acct/s, %.1f min)", done, cfg.Accounts, float64(done)/el, el/60)
				}
			}
		}(w)
	}
	for lo := int64(1); lo <= cfg.Accounts; lo += int64(cfg.Batch) {
		hi := lo + int64(cfg.Batch) - 1
		if hi > cfg.Accounts {
			hi = cfg.Accounts
		}
		select {
		case jobs <- job{lo, hi}:
		case e := <-errCh:
			return e
		}
	}
	close(jobs)
	for w := 0; w < cfg.Workers; w++ {
		<-wg
	}
	select {
	case e := <-errCh:
		return e
	default:
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO bench_meta(k,v) VALUES ('accounts',$1),('slots',$2),('seed',$3),('initial_balance',$4),('loaded_at',now()::text)
		ON CONFLICT (k) DO UPDATE SET v=EXCLUDED.v`, strconv.FormatInt(cfg.Accounts, 10), strconv.Itoa(cfg.Slots), strconv.FormatUint(cfg.Seed, 10), strconv.FormatInt(cfg.InitialBalance, 10)); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `ANALYZE`); err != nil {
		log("analyze warning: %v", err)
	}
	log("load complete in %.1f min", time.Since(start).Minutes())
	return nil
}

func pgLoadRange(ctx context.Context, conn *sql.Conn, cfg LoadConfig, r *rand.Rand, prof, attrs []byte, lo, hi int64) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var sbA, sbL, sbG strings.Builder
	var argsA, argsL, argsG []any
	sbA.WriteString("INSERT INTO accounts(id,username,level,exp,balance,guild_id,profile) VALUES ")
	sbL.WriteString("INSERT INTO leaderboard(account_id,score,matches) VALUES ")
	sbG.WriteString("INSERT INTO guild_members(guild_id,account_id,role,contribution) VALUES ")
	for id := lo; id <= hi; id++ {
		if id > lo {
			sbA.WriteString(",")
			sbL.WriteString(",")
			sbG.WriteString(",")
		}
		fill(prof, r)
		guild := 1 + int(id%int64(cfg.Guilds))
		n := len(argsA)
		fmt.Fprintf(&sbA, "($%d,$%d,$%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5, n+6, n+7)
		argsA = append(argsA, id, fmt.Sprintf("player_%d", id), 1+r.IntN(100), r.Int64N(1_000_000), cfg.InitialBalance, guild, append([]byte(nil), prof...))
		n = len(argsL)
		fmt.Fprintf(&sbL, "($%d,$%d,$%d)", n+1, n+2, n+3)
		argsL = append(argsL, id, r.Int64N(100_000), r.IntN(500))
		n = len(argsG)
		fmt.Fprintf(&sbG, "($%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4)
		argsG = append(argsG, guild, id, r.IntN(4), r.IntN(10000))
	}
	for _, q := range []struct {
		s string
		a []any
	}{{sbA.String(), argsA}, {sbL.String(), argsL}, {sbG.String(), argsG}} {
		if _, err := tx.ExecContext(ctx, q.s, q.a...); err != nil {
			return err
		}
	}
	var sbI strings.Builder
	var argsI []any
	rows := 0
	flush := func() error {
		if rows == 0 {
			return nil
		}
		_, err := tx.ExecContext(ctx, sbI.String(), argsI...)
		sbI.Reset()
		argsI = argsI[:0]
		rows = 0
		return err
	}
	for id := lo; id <= hi; id++ {
		for s := 1; s <= cfg.Slots; s++ {
			if rows == 0 {
				sbI.WriteString("INSERT INTO inventory(account_id,slot,item_id,qty,attrs) VALUES ")
			} else {
				sbI.WriteString(",")
			}
			fill(attrs, r)
			n := len(argsI)
			fmt.Fprintf(&sbI, "($%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5)
			argsI = append(argsI, id, s, 1+r.IntN(5000), 10+r.IntN(90), append([]byte(nil), attrs...))
			rows++
			if rows >= 2000 {
				if err := flush(); err != nil {
					return err
				}
			}
		}
	}
	if err := flush(); err != nil {
		return err
	}
	return tx.Commit()
}

func PGReset(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "TRUNCATE TABLE "+strings.Join(MySQLTables, ", "))
	return err
}

func PGCreateSchema(ctx context.Context, db *sql.DB) error {
	for _, ddl := range PGSchema {
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

func PGCheckInvariants(ctx context.Context, db *sql.DB) (*InvariantReport, error) {
	r := &InvariantReport{}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory WHERE qty<0`).Scan(&r.NegativeInventory); err != nil {
		return nil, err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM accounts WHERE balance<0`).Scan(&r.NegativeBalance); err != nil {
		return nil, err
	}
	var initBal, accounts int64
	if err := db.QueryRowContext(ctx, `SELECT v::bigint FROM bench_meta WHERE k='initial_balance'`).Scan(&initBal); err != nil {
		return nil, err
	}
	if err := db.QueryRowContext(ctx, `SELECT v::bigint FROM bench_meta WHERE k='accounts'`).Scan(&accounts); err != nil {
		return nil, err
	}
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount),0) FROM purchase_ledger`).Scan(&r.LedgerAmountSum); err != nil {
		return nil, err
	}
	var balSum int64
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(balance),0) FROM accounts`).Scan(&balSum); err != nil {
		return nil, err
	}
	r.SpentFromBalances = initBal*accounts - balSum
	r.LedgerMatches = r.SpentFromBalances == r.LedgerAmountSum
	r.Violations = r.NegativeInventory + r.NegativeBalance
	if !r.LedgerMatches {
		r.Violations++
	}
	return r, nil
}
