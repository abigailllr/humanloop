package models

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Credits      int    `json:"credits"`
	Submissions  int    `json:"submissions"`
	ReferralCode string `json:"referral_code,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type Challenge struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Difficulty  string  `json:"difficulty"`
	Submissions int     `json:"submissions"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type Submission struct {
	ID               string   `json:"id"`
	ChallengeID      string   `json:"challenge_id"`
	ChallengeTitle   string   `json:"challenge_title"`
	UserID           string   `json:"user_id"`
	VideoPath        string   `json:"video_path,omitempty"`
	HmdfPath         string   `json:"hmdf_path,omitempty"`
	Status           string   `json:"status"`
	CreditsEarned    int      `json:"credits_earned"`
	QualityScore     float64  `json:"quality_score,omitempty"`
	ExtractorVersion string   `json:"extractor_version,omitempty"`
	ConsentVersion   string   `json:"consent_version,omitempty"`
	VideoHash        string   `json:"-"`
	Approved         bool     `json:"approved"`
	Tags             []string `json:"tags,omitempty"`
	Latitude         float64  `json:"latitude,omitempty"`
	Longitude        float64  `json:"longitude,omitempty"`
	CapturedAt       string   `json:"captured_at,omitempty"`
	CreatedAt        string   `json:"created_at"`
}

type Dataset struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	RobotType   string  `json:"robot_type,omitempty"`
	ChallengeID string  `json:"challenge_id,omitempty"`
	MinQuality  float64 `json:"min_quality,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type BuyerKey struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	DatasetID string `json:"dataset_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

type Webhook struct {
	ID         string `json:"id"`
	DatasetID  string `json:"dataset_id,omitempty"`
	URL        string `json:"url"`
	SecretHash string `json:"-"`
	Active     bool   `json:"active"`
	CreatedAt  string `json:"created_at"`
}
