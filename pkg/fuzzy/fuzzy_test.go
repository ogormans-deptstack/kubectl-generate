package fuzzy

import (
	"testing"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		expected int
	}{
		{"identical strings", "Deployment", "Deployment", 0},
		{"case insensitive match", "deployment", "Deployment", 0},
		{"single transposition", "Deploymnet", "Deployment", 2},
		{"single char insertion", "Deploymentt", "Deployment", 1},
		{"single char deletion", "Deploment", "Deployment", 1},
		{"single char substitution", "Deploymant", "Deployment", 1},
		{"completely different", "Pod", "Ingress", 7},
		{"empty first string", "", "Deployment", 10},
		{"empty second string", "Deployment", "", 10},
		{"both empty", "", "", 0},
		{"prefix match", "Deploy", "Deployment", 4},
		{"short typo", "po", "Pod", 1},
		{"alias not close", "svc", "Service", 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Distance(tc.a, tc.b)
			if got != tc.expected {
				t.Errorf("Distance(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.expected)
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	candidates := []string{
		"Deployment", "DaemonSet", "Pod", "Service", "ConfigMap",
		"Secret", "Job", "CronJob", "Ingress", "StatefulSet",
		"NetworkPolicy", "PersistentVolumeClaim", "HorizontalPodAutoscaler",
	}

	t.Run("exact match returns single result", func(t *testing.T) {
		got := Suggest("Deployment", candidates, 3)
		if len(got) != 1 || got[0] != "Deployment" {
			t.Errorf("expected [Deployment], got %v", got)
		}
	})

	t.Run("case insensitive exact match", func(t *testing.T) {
		got := Suggest("deployment", candidates, 3)
		if len(got) != 1 || got[0] != "Deployment" {
			t.Errorf("expected [Deployment], got %v", got)
		}
	})

	t.Run("close typo returns correct suggestion first", func(t *testing.T) {
		got := Suggest("Deploymnet", candidates, 3)
		if len(got) == 0 || got[0] != "Deployment" {
			t.Errorf("expected Deployment as first suggestion, got %v", got)
		}
	})

	t.Run("Deplyoment typo suggests Deployment", func(t *testing.T) {
		got := Suggest("Deplyoment", candidates, 3)
		if len(got) == 0 || got[0] != "Deployment" {
			t.Errorf("expected Deployment as first suggestion, got %v", got)
		}
	})

	t.Run("Servce typo suggests Service", func(t *testing.T) {
		got := Suggest("Servce", candidates, 3)
		if len(got) == 0 || got[0] != "Service" {
			t.Errorf("expected Service as first suggestion, got %v", got)
		}
	})

	t.Run("completely unrelated input returns empty", func(t *testing.T) {
		got := Suggest("zzzzzzzzzzzzzzz", candidates, 3)
		if len(got) != 0 {
			t.Errorf("expected no suggestions for garbage input, got %v", got)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		got := Suggest("", candidates, 3)
		if len(got) != 0 {
			t.Errorf("expected no suggestions for empty input, got %v", got)
		}
	})

	t.Run("empty candidates returns empty", func(t *testing.T) {
		got := Suggest("Deployment", nil, 3)
		if len(got) != 0 {
			t.Errorf("expected no suggestions from nil candidates, got %v", got)
		}
	})

	t.Run("maxResults limits output", func(t *testing.T) {
		got := Suggest("Set", candidates, 2)
		if len(got) > 2 {
			t.Errorf("expected at most 2 results, got %d: %v", len(got), got)
		}
	})

	t.Run("results sorted by distance", func(t *testing.T) {
		got := Suggest("DaemonSt", candidates, 3)
		if len(got) == 0 {
			t.Fatal("expected at least one suggestion")
		}
		if got[0] != "DaemonSet" {
			t.Errorf("expected DaemonSet as closest match, got %v", got)
		}
	})

	t.Run("single character input only matches very short kinds", func(t *testing.T) {
		got := Suggest("P", candidates, 3)
		// threshold is max(len(input)/2, 3) => for "P" that's max(0, 3)=3
		// "Pod" has distance 2 from "P" (case insensitive: "p" vs "pod" = 2)
		// so Pod should be suggested
		if len(got) > 0 && got[0] != "Pod" {
			t.Errorf("for single char 'P', expected Pod if anything, got %v", got)
		}
	})
}
