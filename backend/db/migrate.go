package db

import "context"

func Migrate(ctx context.Context) error {
	_, err := Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id          TEXT PRIMARY KEY,
			email       TEXT NOT NULL,
			name        TEXT NOT NULL,
			credits     INT  NOT NULL DEFAULT 0,
			submissions INT  NOT NULL DEFAULT 0,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS challenges (
			id          TEXT PRIMARY KEY,
			title       TEXT NOT NULL,
			description TEXT NOT NULL,
			difficulty  TEXT NOT NULL DEFAULT 'Easy',
			submissions INT  NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS submissions (
			id             TEXT PRIMARY KEY,
			challenge_id   TEXT NOT NULL REFERENCES challenges(id),
			user_id        TEXT NOT NULL REFERENCES users(id),
			video_path     TEXT NOT NULL,
			hmdf_path      TEXT,
			status         TEXT NOT NULL DEFAULT 'pending',
			credits_earned INT  NOT NULL DEFAULT 0,
			latitude       FLOAT,
			longitude      FLOAT,
			captured_at    TIMESTAMPTZ,
			synthetic      BOOLEAN NOT NULL DEFAULT FALSE,
			robot          TEXT NOT NULL DEFAULT 'generic_bimanual',
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS credit_transactions (
			id             BIGSERIAL PRIMARY KEY,
			user_id        TEXT NOT NULL REFERENCES users(id),
			submission_id  TEXT REFERENCES submissions(id),
			amount         INT  NOT NULL,
			reason         TEXT NOT NULL,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS submissions_user_id      ON submissions(user_id);
		CREATE INDEX IF NOT EXISTS submissions_challenge_id ON submissions(challenge_id);
		CREATE INDEX IF NOT EXISTS submissions_status       ON submissions(status);
		CREATE INDEX IF NOT EXISTS submissions_synthetic    ON submissions(synthetic) WHERE synthetic = TRUE;
		CREATE INDEX IF NOT EXISTS credit_tx_user_id        ON credit_transactions(user_id);

		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS synthetic          BOOLEAN  NOT NULL DEFAULT FALSE;
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS robot             TEXT     NOT NULL DEFAULT 'generic_bimanual';
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS quality_score     FLOAT    NOT NULL DEFAULT 0;
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS extractor_version TEXT     NOT NULL DEFAULT '';
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS consent_version   TEXT     NOT NULL DEFAULT '1.0';
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS video_hash        TEXT     NOT NULL DEFAULT '';
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS approved          BOOLEAN  NOT NULL DEFAULT TRUE;
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS tags              TEXT[]   NOT NULL DEFAULT '{}';

		CREATE TABLE IF NOT EXISTS datasets (
			id           TEXT PRIMARY KEY,
			title        TEXT NOT NULL,
			description  TEXT NOT NULL DEFAULT '',
			robot_type   TEXT NOT NULL DEFAULT '',
			challenge_id TEXT NOT NULL DEFAULT '',
			min_quality  FLOAT NOT NULL DEFAULT 0,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS buyer_keys (
			id         TEXT PRIMARY KEY,
			key_hash   TEXT NOT NULL UNIQUE,
			label      TEXT NOT NULL,
			dataset_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS submissions_video_hash ON submissions(video_hash) WHERE video_hash != '';
		CREATE INDEX IF NOT EXISTS submissions_approved   ON submissions(approved)   WHERE approved = FALSE;
		CREATE INDEX IF NOT EXISTS submissions_quality    ON submissions(quality_score);
	`)
	return err
}
