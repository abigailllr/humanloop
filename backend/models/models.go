package models

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Credits     int    `json:"credits"`
	Submissions int    `json:"submissions"`
	CreatedAt   string `json:"created_at"`
}

type Challenge struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	Submissions int    `json:"submissions"`
}

type Submission struct {
	ID             string  `json:"id"`
	ChallengeID    string  `json:"challenge_id"`
	ChallengeTitle string  `json:"challenge_title"`
	UserID         string  `json:"user_id"`
	VideoPath      string  `json:"video_path"`
	Status         string  `json:"status"`
	CreditsEarned  int     `json:"credits_earned"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	CapturedAt     string  `json:"captured_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}
