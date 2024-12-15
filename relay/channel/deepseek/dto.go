package deepseek

import (
	"one-api/dto"
)

type ChatCompletionsStreamResponse struct {
	Id                string                                    `json:"id"`
	Object            string                                    `json:"object"`
	Created           int64                                     `json:"created"`
	Model             string                                    `json:"model"`
	SystemFingerprint *string                                   `json:"system_fingerprint"`
	Choices           []dto.ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage             *Usage                                    `json:"usage"`
}

type PromptUseCacheTokenType struct {
	ChatCompletionsStreamResponse
}

type DeepseekTextResponse struct {
	Id      string                         `json:"id"`
	Model   string                         `json:"model"`
	Object  string                         `json:"object"`
	Created int64                          `json:"created"`
	Choices []dto.OpenAITextResponseChoice `json:"choices"`
	Usage   *Usage                         `json:"usage"`
}
type Usage struct {
	PromptTokens           int                    `json:"prompt_tokens"`
	CompletionTokens       int                    `json:"completion_tokens"`
	TotalTokens            int                    `json:"total_tokens"`
	PromptTokensDetails    dto.InputTokenDetails  `json:"prompt_tokens_details"`
	CompletionTokenDetails dto.OutputTokenDetails `json:"completion_tokens_details"`
	PromptCacheHitTokens   int                    `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens  int                    `json:"prompt_cache_miss_tokens"`
}
