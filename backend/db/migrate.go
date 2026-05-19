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
			duration       FLOAT,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS submissions_user_id ON submissions(user_id);
		CREATE INDEX IF NOT EXISTS submissions_challenge_id ON submissions(challenge_id);
		CREATE INDEX IF NOT EXISTS submissions_status ON submissions(status);

		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS credits_earned INT NOT NULL DEFAULT 0;
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS latitude FLOAT;
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS longitude FLOAT;
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS captured_at TIMESTAMPTZ;
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS synthetic BOOLEAN NOT NULL DEFAULT FALSE;
	`)
	return err
}
