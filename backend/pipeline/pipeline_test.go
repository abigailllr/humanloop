package pipeline

import (
	"testing"
)

func makeFrames(n, withMotor int, fps float64) map[string]any {
	frames := make([]any, n)
	for i := range frames {
		f := map[string]any{"t": float64(i) / fps}
		if i < withMotor {
			f["motor_state"] = map[string]any{"q": []any{0.1, 0.2}}
		}
		frames[i] = f
	}
	return map[string]any{"frames": frames, "fps": fps}
}

func TestQualityScore_Empty(t *testing.T) {
	if got := qualityScore(map[string]any{}); got != 0 {
		t.Fatalf("want 0, got %f", got)
	}
}

func TestQualityScore_Perfect(t *testing.T) {
	rec := makeFrames(300, 300, 30)
	if got := qualityScore(rec); got != 1.0 {
		t.Fatalf("want 1.0, got %f", got)
	}
}

func TestQualityScore_PartialMotor(t *testing.T) {
	rec := makeFrames(300, 150, 30)
	got := qualityScore(rec)
	want := (0.5 + 1.0 + 1.0) / 3.0
	if abs(got-want) > 1e-9 {
		t.Fatalf("want %f, got %f", want, got)
	}
}

func TestQualityScore_LowFPS(t *testing.T) {
	rec := makeFrames(300, 300, 15)
	got := qualityScore(rec)
	want := (1.0 + 0.5 + 1.0) / 3.0
	if abs(got-want) > 1e-9 {
		t.Fatalf("want %f, got %f", want, got)
	}
}

func TestQualityScore_FPSCappedAtOne(t *testing.T) {
	rec := makeFrames(300, 300, 60)
	got := qualityScore(rec)
	if got != 1.0 {
		t.Fatalf("fps > 30 should still cap at 1.0, got %f", got)
	}
}

func TestQualityScore_ShortDuration(t *testing.T) {
	rec := makeFrames(30, 30, 30)
	got := qualityScore(rec)
	want := (1.0 + 1.0 + 30.0/300.0) / 3.0
	if abs(got-want) > 1e-9 {
		t.Fatalf("want %f, got %f", want, got)
	}
}

func TestIsSynthetic_Missing(t *testing.T) {
	if isSynthetic(map[string]any{}) {
		t.Fatal("expected false when synthetic_detection missing")
	}
}

func TestIsSynthetic_False(t *testing.T) {
	rec := map[string]any{
		"synthetic_detection": map[string]any{"synthetic": false},
	}
	if isSynthetic(rec) {
		t.Fatal("expected false")
	}
}

func TestIsSynthetic_True(t *testing.T) {
	rec := map[string]any{
		"synthetic_detection": map[string]any{"synthetic": true},
	}
	if !isSynthetic(rec) {
		t.Fatal("expected true")
	}
}

func TestCreditsForDifficulty(t *testing.T) {
	cases := []struct {
		d    string
		want int
	}{
		{"Hard", 20},
		{"Medium", 15},
		{"Easy", 10},
		{"", 10},
		{"Unknown", 10},
	}
	for _, c := range cases {
		if got := creditsForDifficulty(c.d); got != c.want {
			t.Errorf("difficulty %q: want %d, got %d", c.d, c.want, got)
		}
	}
}

func TestPipelineShutdown(t *testing.T) {
	p := New(2, t.TempDir(), nil)
	p.Shutdown()
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
