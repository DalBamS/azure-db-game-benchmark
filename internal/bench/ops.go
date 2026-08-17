package bench

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

// Op is one game operation executed on a dedicated connection.
type Op struct {
	Name  string
	Write bool
	Run   func(ctx context.Context, conn *sql.Conn, w *Worker) error
}

// Worker-local state passed to ops.
type Worker struct {
	ID    int
	Rng   *rand.Rand
	Keys  *KeyPicker
	Buf   []byte // scratch for random payloads
	Slots int    // inventory slots per account
	Now   func() time.Time
}

func (w *Worker) randBytes(n int) []byte {
	if cap(w.Buf) < n {
		w.Buf = make([]byte, n)
	}
	b := w.Buf[:n]
	// ChaCha8-backed rand is incompressible enough for storage-size purposes.
	for i := 0; i+8 <= n; i += 8 {
		binary.LittleEndian.PutUint64(b[i:], w.Rng.Uint64())
	}
	for i := n &^ 7; i < n; i++ {
		b[i] = byte(w.Rng.Uint64())
	}
	return b
}

func uuidBytes(r *rand.Rand) []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, r.Uint64())
	binary.LittleEndian.PutUint64(b[8:], r.Uint64())
	return b
}

var ErrInvariant = errors.New("invariant violation")

// ---- MySQL operations -------------------------------------------------------

func opProfileRead() Op {
	return Op{Name: "profile_read", Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		var level int
		var exp, bal int64
		var profile []byte
		return c.QueryRowContext(ctx, `SELECT level, exp, balance, profile FROM accounts WHERE id=?`, id).Scan(&level, &exp, &bal, &profile)
	}}
}

func opInventoryRead() Op {
	return Op{Name: "inventory_read", Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		rows, err := c.QueryContext(ctx, `SELECT slot, item_id, qty, attrs FROM inventory WHERE account_id=?`, id)
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

func opLeaderboardRead() Op {
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

func opGuildRead() Op {
	return Op{Name: "guild_read", Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		rows, err := c.QueryContext(ctx, `SELECT g.guild_id, g.name, g.rating, gm.account_id, gm.role, gm.contribution
			FROM accounts a JOIN guilds g ON g.guild_id=a.guild_id JOIN guild_members gm ON gm.guild_id=g.guild_id
			WHERE a.id=? LIMIT 50`, id)
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

func opSessionUpsert() Op {
	return Op{Name: "session_upsert", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		_, err := c.ExecContext(ctx, `INSERT INTO game_sessions(account_id, session_token, client_version, last_seen)
			VALUES (?, ?, '1.0.0', NOW(6)) ON DUPLICATE KEY UPDATE last_seen=VALUES(last_seen), login_count=login_count+1`, id, uuidBytes(w.Rng))
		return err
	}}
}

func opInventoryUpdate() Op {
	return Op{Name: "inventory_update", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		slot := 1 + w.Rng.IntN(w.Slots)
		tx, err := c.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var qty int
		if err := tx.QueryRowContext(ctx, `SELECT qty FROM inventory WHERE account_id=? AND slot=? FOR UPDATE`, id, slot).Scan(&qty); err != nil {
			return err
		}
		delta := 1 + w.Rng.IntN(5)
		if qty-delta < 0 {
			delta = 0 // never drive quantity negative (application-level invariant)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE inventory SET qty=qty-?, version=version+1, attrs=? WHERE account_id=? AND slot=?`,
			delta, w.randBytes(256), id, slot); err != nil {
			return err
		}
		return tx.Commit()
	}}
}

func opMatchResult() Op {
	return Op{Name: "match_result", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		score := w.Rng.IntN(1000)
		tx, err := c.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if _, err := tx.ExecContext(ctx, `INSERT INTO match_results(account_id, score, duration_ms, payload) VALUES (?,?,?,?)`,
			id, score, 30000+w.Rng.IntN(600000), w.randBytes(128)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE leaderboard SET score=score+?, matches=matches+1 WHERE account_id=?`, score, id); err != nil {
			return err
		}
		return tx.Commit()
	}}
}

func opPurchase() Op {
	return Op{Name: "purchase", Write: true, Run: func(ctx context.Context, c *sql.Conn, w *Worker) error {
		id := w.Keys.Pick()
		item := 1 + w.Rng.IntN(5000)
		amount := int64(10 + w.Rng.IntN(90))
		tx, err := c.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		res, err := tx.ExecContext(ctx, `UPDATE accounts SET balance=balance-? WHERE id=? AND balance>=?`, amount, id, amount)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			// insufficient balance: business no-op, still a successful transaction
			return tx.Commit()
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO purchase_ledger(request_id, account_id, item_id, amount) VALUES (?,?,?,?)`,
			uuidBytes(w.Rng), id, item, amount); err != nil {
			return err
		}
		slot := 1 + w.Rng.IntN(w.Slots)
		if _, err := tx.ExecContext(ctx, `UPDATE inventory SET qty=qty+1, version=version+1 WHERE account_id=? AND slot=?`, id, slot); err != nil {
			return err
		}
		return tx.Commit()
	}}
}

// login = profile read + session upsert + inventory read (S3 burst unit)
func opLogin() Op {
	pr, su, ir := opProfileRead(), opSessionUpsert(), opInventoryRead()
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

// Mix is a weighted list of ops.
type Mix struct {
	Ops     []Op
	Weights []int
	total   int
}

func (m *Mix) Pick(r *rand.Rand) *Op {
	if m.total == 0 {
		for _, w := range m.Weights {
			m.total += w
		}
	}
	x := r.IntN(m.total)
	for i, w := range m.Weights {
		if x < w {
			return &m.Ops[i]
		}
		x -= w
	}
	return &m.Ops[len(m.Ops)-1]
}

func (m *Mix) Describe() map[string]int {
	out := map[string]int{}
	for i, op := range m.Ops {
		out[op.Name] = m.Weights[i]
	}
	return out
}

func ScenarioMix(name string) (*Mix, error) {
	switch name {
	case "S1", "s1", "normal":
		return &Mix{
			Ops:     []Op{opProfileRead(), opInventoryRead(), opLeaderboardRead(), opGuildRead(), opSessionUpsert(), opInventoryUpdate(), opMatchResult(), opPurchase()},
			Weights: []int{30, 20, 10, 5, 10, 12, 8, 5},
		}, nil
	case "S2", "s2", "write":
		return &Mix{
			Ops:     []Op{opMatchResult(), opInventoryUpdate(), opPurchase(), opSessionUpsert(), opProfileRead(), opLeaderboardRead()},
			Weights: []int{30, 25, 15, 10, 15, 5},
		}, nil
	case "S3", "s3", "login":
		return &Mix{
			Ops:     []Op{opLogin(), opProfileRead(), opInventoryRead(), opInventoryUpdate(), opMatchResult()},
			Weights: []int{40, 20, 15, 15, 10},
		}, nil
	case "S4", "s4", "hotspot":
		m, _ := ScenarioMix("S2")
		return m, nil
	case "read":
		return &Mix{Ops: []Op{opProfileRead(), opInventoryRead()}, Weights: []int{50, 50}}, nil
	}
	return nil, fmt.Errorf("unknown scenario %q", name)
}
