package bench

// Game OLTP schema. Payload columns are random bytes (incompressible) so that
// on-disk size ~= logical size and G1 (dataset >= 2x buffer pool) is meaningful.

var MySQLSchema = []string{
	`CREATE TABLE IF NOT EXISTS accounts (
		id BIGINT NOT NULL PRIMARY KEY,
		username VARCHAR(32) NOT NULL,
		level INT NOT NULL,
		exp BIGINT NOT NULL,
		balance BIGINT NOT NULL,
		guild_id INT NOT NULL,
		profile VARBINARY(2048) NOT NULL,
		updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		KEY ix_accounts_guild (guild_id)
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS inventory (
		account_id BIGINT NOT NULL,
		slot SMALLINT NOT NULL,
		item_id INT NOT NULL,
		qty INT NOT NULL,
		version INT NOT NULL DEFAULT 0,
		attrs VARBINARY(512) NOT NULL,
		PRIMARY KEY (account_id, slot)
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS game_sessions (
		account_id BIGINT NOT NULL PRIMARY KEY,
		session_token BINARY(16) NOT NULL,
		client_version VARCHAR(16) NOT NULL,
		last_seen DATETIME(6) NOT NULL,
		login_count INT NOT NULL DEFAULT 1
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS purchase_ledger (
		request_id BINARY(16) NOT NULL PRIMARY KEY,
		account_id BIGINT NOT NULL,
		item_id INT NOT NULL,
		amount BIGINT NOT NULL,
		created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		KEY ix_ledger_account (account_id, created_at)
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS match_results (
		match_id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		account_id BIGINT NOT NULL,
		score INT NOT NULL,
		duration_ms INT NOT NULL,
		payload VARBINARY(256) NOT NULL,
		created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		KEY ix_match_account (account_id, created_at)
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS leaderboard (
		account_id BIGINT NOT NULL PRIMARY KEY,
		score BIGINT NOT NULL,
		matches INT NOT NULL,
		updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		KEY ix_lb_score (score DESC)
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS guilds (
		guild_id INT NOT NULL PRIMARY KEY,
		name VARCHAR(64) NOT NULL,
		rating INT NOT NULL,
		notice VARBINARY(1024) NOT NULL
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS guild_members (
		guild_id INT NOT NULL,
		account_id BIGINT NOT NULL,
		role TINYINT NOT NULL,
		contribution INT NOT NULL,
		PRIMARY KEY (guild_id, account_id)
	) ENGINE=InnoDB`,
	`CREATE TABLE IF NOT EXISTS bench_meta (
		k VARCHAR(64) NOT NULL PRIMARY KEY,
		v VARCHAR(255) NOT NULL
	) ENGINE=InnoDB`,
}

var MySQLTables = []string{"accounts", "inventory", "game_sessions", "purchase_ledger", "match_results", "leaderboard", "guilds", "guild_members", "bench_meta"}
