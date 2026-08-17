package bench

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LoadConfig struct {
	Accounts       int64
	Slots          int
	ProfileBytes   int
	AttrsBytes     int
	Guilds         int
	Workers        int
	Batch          int
	Seed           uint64
	InitialBalance int64
}

// Load populates the schema in parallel. It is idempotent per range only when the tables
// are empty; callers should TRUNCATE first (see Reset).
func Load(ctx context.Context, db *sql.DB, cfg LoadConfig, log func(string, ...any)) error {
	if cfg.Workers <= 0 {
		cfg.Workers = 16
	}
	if cfg.Batch <= 0 {
		cfg.Batch = 500
	}
	start := time.Now()
	// guilds first
	{
		var sb strings.Builder
		args := []any{}
		r := rand.New(rand.NewPCG(cfg.Seed, 1))
		buf := make([]byte, 1024)
		for g := 1; g <= cfg.Guilds; g++ {
			if sb.Len() == 0 {
				sb.WriteString("INSERT INTO guilds(guild_id,name,rating,notice) VALUES ")
			} else {
				sb.WriteString(",")
			}
			sb.WriteString("(?,?,?,?)")
			fill(buf, r)
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
	// account ranges
	var next atomic.Int64
	next.Store(1)
	var done atomic.Int64
	var wg sync.WaitGroup
	errCh := make(chan error, cfg.Workers)
	chunk := int64(cfg.Batch)
	for w := 0; w < cfg.Workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewPCG(cfg.Seed+uint64(id)*31337, uint64(id)))
			prof := make([]byte, cfg.ProfileBytes)
			attrs := make([]byte, cfg.AttrsBytes)
			conn, err := db.Conn(ctx)
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()
			for {
				lo := next.Add(chunk) - chunk
				if lo > cfg.Accounts {
					return
				}
				hi := lo + chunk - 1
				if hi > cfg.Accounts {
					hi = cfg.Accounts
				}
				if err := loadRange(ctx, conn, cfg, r, prof, attrs, lo, hi); err != nil {
					errCh <- fmt.Errorf("range %d-%d: %w", lo, hi, err)
					return
				}
				d := done.Add(hi - lo + 1)
				if (d/chunk)%200 == 0 {
					el := time.Since(start).Seconds()
					log("loaded %d/%d accounts (%.0f acct/s, %.1f min elapsed)", d, cfg.Accounts, float64(d)/el, el/60)
				}
			}
		}(w)
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		if e != nil {
			return e
		}
	}
	if _, err := db.ExecContext(ctx, `REPLACE INTO bench_meta(k,v) VALUES ('accounts',?),('slots',?),('seed',?),('initial_balance',?),('loaded_at',UTC_TIMESTAMP(6))`,
		cfg.Accounts, cfg.Slots, cfg.Seed, cfg.InitialBalance); err != nil {
		return err
	}
	log("load complete in %.1f min", time.Since(start).Minutes())
	return nil
}

func fill(b []byte, r *rand.Rand) {
	i := 0
	for ; i+8 <= len(b); i += 8 {
		binary.LittleEndian.PutUint64(b[i:], r.Uint64())
	}
	for ; i < len(b); i++ {
		b[i] = byte(r.Uint64())
	}
}

func loadRange(ctx context.Context, conn *sql.Conn, cfg LoadConfig, r *rand.Rand, prof, attrs []byte, lo, hi int64) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// accounts + leaderboard + guild_members
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
		sbA.WriteString("(?,?,?,?,?,?,?)")
		argsA = append(argsA, id, fmt.Sprintf("player_%d", id), 1+r.IntN(100), r.Int64N(1_000_000), cfg.InitialBalance, guild, append([]byte(nil), prof...))
		sbL.WriteString("(?,?,?)")
		argsL = append(argsL, id, r.Int64N(100_000), r.IntN(500))
		sbG.WriteString("(?,?,?,?)")
		argsG = append(argsG, guild, id, r.IntN(4), r.IntN(10000))
	}
	if _, err := tx.ExecContext(ctx, sbA.String(), argsA...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, sbL.String(), argsL...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, sbG.String(), argsG...); err != nil {
		return err
	}
	// inventory: slots per account, in sub-batches to keep packet size sane
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
			sbI.WriteString("(?,?,?,?,?)")
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

// Reset truncates all workload tables.
func Reset(ctx context.Context, db *sql.DB) error {
	for _, t := range MySQLTables {
		if _, err := db.ExecContext(ctx, "TRUNCATE TABLE "+t); err != nil {
			return fmt.Errorf("truncate %s: %w", t, err)
		}
	}
	return nil
}

func CreateSchema(ctx context.Context, db *sql.DB) error {
	for _, ddl := range MySQLSchema {
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

// InvariantCheck verifies application-level invariants (G9): no negative inventory,
// balances never below zero, and ledger sum == total spent.
type InvariantReport struct {
	NegativeInventory int64 `json:"negative_inventory_rows"`
	NegativeBalance   int64 `json:"negative_balance_rows"`
	LedgerAmountSum   int64 `json:"ledger_amount_sum"`
	SpentFromBalances int64 `json:"spent_from_balances"`
	LedgerMatches     bool  `json:"ledger_matches_balances"`
	Violations        int64 `json:"violations"`
}

func CheckInvariants(ctx context.Context, db *sql.DB) (*InvariantReport, error) {
	r := &InvariantReport{}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory WHERE qty<0`).Scan(&r.NegativeInventory); err != nil {
		return nil, err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM accounts WHERE balance<0`).Scan(&r.NegativeBalance); err != nil {
		return nil, err
	}
	var initBal, accounts int64
	if err := db.QueryRowContext(ctx, `SELECT CAST(v AS SIGNED) FROM bench_meta WHERE k='initial_balance'`).Scan(&initBal); err != nil {
		return nil, err
	}
	if err := db.QueryRowContext(ctx, `SELECT CAST(v AS SIGNED) FROM bench_meta WHERE k='accounts'`).Scan(&accounts); err != nil {
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
