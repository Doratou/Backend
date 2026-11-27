package entity

// The reward key contains spaces, which purchase idempotency keys cannot contain.
const ProfileCompletionRewardKey = "profile completion reward"

type ProfileCompletion struct {
	FullName          bool
	ProfilePhoto      bool
	VerifiedEmail     bool
	VerifiedPhone     bool
	TelegramConnected bool
	RewardTokens      int64
	RewardGranted     bool
}
