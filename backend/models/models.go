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
	Valid          bool    `json:"valid"`
	CreditsEarned  int     `json:"credits_earned"`
	Duration       float64 `json:"duration"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	CapturedAt     string  `json:"captured_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type Landmark struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	V float64 `json:"v,omitempty"`
}

type LandmarkVelocity struct {
	VX float64 `json:"vx"`
	VY float64 `json:"vy"`
	VZ float64 `json:"vz"`
}

type DetectedObject struct {
	Label      string    `json:"label"`
	Confidence float64   `json:"confidence"`
	BBox       []float64 `json:"bbox"`
}

type Contact struct {
	Hand        int     `json:"hand"`
	ObjectLabel string  `json:"object_label"`
	Confidence  float64 `json:"confidence"`
}

type ContactEvent struct {
	Hand        int     `json:"hand"`
	ObjectLabel string  `json:"object_label"`
	TStart      float64 `json:"t_start"`
	TEnd        float64 `json:"t_end"`
}

type HMDFFrame struct {
	T          float64            `json:"t"`
	Pose       []Landmark         `json:"pose"`
	PoseVel    []LandmarkVelocity `json:"pose_vel,omitempty"`
	Hands      [][]Landmark       `json:"hands"`
	HandVel    [][]LandmarkVelocity `json:"hand_vel,omitempty"`
	HandLabels []string           `json:"hand_labels,omitempty"`
	Objects    []DetectedObject   `json:"objects"`
	Contacts   []Contact          `json:"contacts"`
}

type HMDFStats struct {
	Bimanual       bool     `json:"bimanual"`
	ObjectsTouched []string `json:"objects_touched"`
	DominantHand   string   `json:"dominant_hand,omitempty"`
}

type HMDFMetadata struct {
	TaskType        string `json:"task_type"`
	CoordinateSpace string `json:"coordinate_space"`
	PoseLandmarks   int    `json:"pose_landmarks"`
	HandLandmarks   int    `json:"hand_landmarks"`
	VisionModel     string `json:"vision_model"`
}

type ActionPhase struct {
	Label    string  `json:"label"`
	StartPct float64 `json:"start_pct"`
	EndPct   float64 `json:"end_pct"`
}

type HMDFValidation struct {
	Valid      bool          `json:"valid"`
	Confidence float64       `json:"confidence"`
	Scene      string        `json:"scene"`
	Reason     string        `json:"reason"`
	Phases     []ActionPhase `json:"phases,omitempty"`
	Skipped    bool          `json:"skipped,omitempty"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type HMDFRecord struct {
	HMDFVersion    string          `json:"hmdf_version"`
	Source         string          `json:"source"`
	SubmissionID   string          `json:"submission_id"`
	ChallengeID    string          `json:"challenge_id"`
	ChallengeTitle string          `json:"challenge_title"`
	UserID         string          `json:"user_id"`
	Location       *Location       `json:"location,omitempty"`
	CapturedAt     string          `json:"captured_at,omitempty"`
	FPS            float64         `json:"fps"`
	FrameCount     int             `json:"frame_count"`
	Frames         []HMDFFrame     `json:"frames"`
	ContactEvents  []ContactEvent  `json:"contact_events,omitempty"`
	Stats          *HMDFStats      `json:"stats,omitempty"`
	Validation     *HMDFValidation `json:"validation,omitempty"`
	Metadata       HMDFMetadata    `json:"metadata"`
}
