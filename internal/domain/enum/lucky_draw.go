package enum

type LuckyDrawStatus string

const (
	LuckyDrawStatusDraft     LuckyDrawStatus = "DRAFT"
	LuckyDrawStatusCompleted LuckyDrawStatus = "COMPLETED"
)

type LuckyDrawType string

const (
	LuckyDrawTypeEqual    LuckyDrawType = "EQUAL"
	LuckyDrawTypeWeighted LuckyDrawType = "WEIGHTED"
)

type LuckyDrawCandidateType string

const (
	LuckyDrawCandidateTypeText LuckyDrawCandidateType = "TEXT"
	LuckyDrawCandidateTypeUser LuckyDrawCandidateType = "USER"
)
