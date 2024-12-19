package deepseek

import (
	"one-api/dto"
)

type DeepseekTextResponse struct {
	Id      string                         `json:"id"`
	Model   string                         `json:"model"`
	Object  string                         `json:"object"`
	Created int64                          `json:"created"`
	Choices []dto.OpenAITextResponseChoice `json:"choices"`
	Usage   *dto.Usage                     `json:"usage"`
}
