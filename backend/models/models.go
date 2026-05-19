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

type HMDFMetadata struct {
	TaskType        string `json:"task_type"`
	CoordinateSpace string `json:"coordinate_space"`
	PoseLandmarks   int    `json:"pose_landmarks"`
	HandLandmarks   int    `json:"hand_landmarks"`
}

type HMDFRecord struct {
	HMDFVersion    string        `json:"hmdf_version"`
	Source         string        `json:"source"`
	SubmissionID   string        `json:"submission_id"`
	ChallengeID    string        `json:"challenge_id"`
	ChallengeTitle string        `json:"challenge_title"`
	UserID         string        `json:"user_id"`
	FPS            float64       `json:"fps"`
	FrameCount     int           `json:"frame_count"`
	Frames         []PoseFrame   `json:"frames"`
	Metadata       HMDFMetadata  `json:"metadata"`
}
