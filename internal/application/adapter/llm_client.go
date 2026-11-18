package adapter

import "context"

type LLMClient interface {
	Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}