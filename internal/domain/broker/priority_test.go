package broker

import "testing"

func TestCalculatePriority_Fresh(t *testing.T) {
	tests := []struct {
		name     string
		urgency  Urgency
		expected uint8
	}{
		{name: "normal fresh", urgency: UrgencyNormal, expected: 1},
		{name: "standard fresh", urgency: UrgencyStandard, expected: 4},
		{name: "realtime fresh", urgency: UrgencyRealtime, expected: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePriority(tt.urgency, 0, 3)
			if got != tt.expected {
				t.Errorf("CalculatePriority(%v, 0, 3) = %d, expected %d", tt.urgency, got, tt.expected)
			}
		})
	}
}

func TestCalculatePriority_RetriesWithinBands(t *testing.T) {
	maxRetriesList := []int{1, 3, 5, 10}

	for _, max := range maxRetriesList {
		for retry := 1; retry <= max; retry++ {
			normalPriority := CalculatePriority(UrgencyNormal, retry, max)
			standardPriority := CalculatePriority(UrgencyStandard, retry, max)
			realtimePriority := CalculatePriority(UrgencyRealtime, retry, max)

			if normalPriority < 2 || normalPriority > 3 {
				t.Errorf("normal priority %d outside [2, 3] for retry %d/%d", normalPriority, retry, max)
			}

			if standardPriority < 5 || standardPriority > 6 {
				t.Errorf("standard priority %d outside [5, 6] for retry %d/%d", standardPriority, retry, max)
			}

			if realtimePriority < 8 || realtimePriority > 10 {
				t.Errorf("realtime priority %d outside [8, 10] for retry %d/%d", realtimePriority, retry, max)
			}

			// Invariant 1: Normal retried must NEVER exceed Standard fresh (4)
			if normalPriority >= CalculatePriority(UrgencyStandard, 0, max) {
				t.Errorf("normal retried (%d) must be strictly less than standard fresh (4)", normalPriority)
			}

			// Invariant 2: Standard retried must NEVER exceed Realtime fresh (7)
			if standardPriority >= CalculatePriority(UrgencyRealtime, 0, max) {
				t.Errorf("standard retried (%d) must be strictly less than realtime fresh (7)", standardPriority)
			}

			// Invariant 3: Realtime retried must be strictly greater than Realtime fresh (7)
			if realtimePriority <= CalculatePriority(UrgencyRealtime, 0, max) {
				t.Errorf("realtime retried (%d) must be strictly greater than realtime fresh (7)", realtimePriority)
			}
		}
	}
}

func TestCalculatePriority_EdgeCases(t *testing.T) {
	// Negative retry should act like 0
	if got := CalculatePriority(UrgencyRealtime, -1, 3); got != 7 {
		t.Errorf("expected 7 for negative retry, got %d", got)
	}

	// Zero or negative maxRetries
	if got := CalculatePriority(UrgencyNormal, 1, 0); got != 3 {
		t.Errorf("expected 3 for zero maxRetries, got %d", got)
	}

	// Retry exceeding maxRetries should clamp to maxRetries
	if got := CalculatePriority(UrgencyRealtime, 10, 3); got != 10 {
		t.Errorf("expected 10 for clamped max retry, got %d", got)
	}
}
