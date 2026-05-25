package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func generateReferralCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func EnsureReferralCode(ctx context.Context, userID string) (string, error) {
	code, err := generateReferralCode()
	if err != nil {
		return "", err
	}
	var result string
	err = Pool.QueryRow(ctx, `
		UPDATE users SET referral_code = COALESCE(referral_code, $2)
		WHERE id = $1
		RETURNING referral_code
	`, userID, code).Scan(&result)
	return result, err
}

func GetReferralCode(ctx context.Context, userID string) (string, error) {
	var code string
	err := Pool.QueryRow(ctx, `SELECT referral_code FROM users WHERE id = $1`, userID).Scan(&code)
	return code, err
}

type ReferralStats struct {
	Code       string `json:"code"`
	TotalRefer int    `json:"total_referrals"`
}

func GetReferralStats(ctx context.Context, userID string) (ReferralStats, error) {
	var s ReferralStats
	err := Pool.QueryRow(ctx, `
		SELECT COALESCE(referral_code,''), (SELECT COUNT(*) FROM referrals WHERE referrer_id=$1)
		FROM users WHERE id=$1
	`, userID).Scan(&s.Code, &s.TotalRefer)
	return s, err
}

func RedeemReferral(ctx context.Context, refereeID, code string) error {
	var referrerID string
	err := Pool.QueryRow(ctx, `SELECT id FROM users WHERE referral_code=$1`, code).Scan(&referrerID)
	if err != nil {
		return errors.New("invalid code")
	}
	if referrerID == refereeID {
		return errors.New("cannot redeem own code")
	}
	_, err = Pool.Exec(ctx, `INSERT INTO referrals (referrer_id, referee_id) VALUES ($1, $2)`, referrerID, refereeID)
	if err != nil {
		return errors.New("already redeemed")
	}
	Pool.Exec(ctx, `UPDATE users SET credits = credits + 10 WHERE id = $1`, referrerID)
	Pool.Exec(ctx, `UPDATE users SET credits = credits + 10 WHERE id = $1`, refereeID)
	LogCreditTransaction(ctx, referrerID, "", "referral_bonus", 10)
	LogCreditTransaction(ctx, refereeID, "", "referral_bonus", 10)
	return nil
}
