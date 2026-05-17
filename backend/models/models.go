package models

type Challenge struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Submissions int    `json:"submissions"`
}

type Submission struct {
	ID          string  `json:"id"`
	ChallengeID string  `json:"challenge_id"`
	UserID      string  `json:"user_id"`
	VideoPath   string  `json:"video_path"`
	Valid       bool    `json:"valid"`
	Duration    float64 `json:"duration"`
	CreatedAt   string  `json:"created_at"`
}

type PoseFrame struct {
	Pose  []Landmark   `json:"pose"`
	Hands [][]Landmark `json:"hands"`
}

type Landmark struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	V float64 `json:"v,omitempty"`
}

type ExtractedData struct {
	SubmissionID string      `json:"submission_id"`
	ChallengeID  string      `json:"challenge_id"`
	FPS          float64     `json:"fps"`
	FrameCount   int         `json:"frame_count"`
	Frames       []PoseFrame `json:"frames"`
}
