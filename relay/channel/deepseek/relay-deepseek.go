package deepseek

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func deepseekStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var usage dto.Usage
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	dataChan := make(chan string)
	stopChan := make(chan bool)
	defer close(stopChan)

	// 添加超时处理
	ticker := time.NewTicker(time.Duration(constant.StreamingTimeout) * time.Second)
	defer ticker.Stop()

	var streamItems []string // 存储所有流式响应
	var responseTextBuilder strings.Builder
	var toolCount int
	var mu sync.Mutex

	// 启动goroutine来读取响应数据
	go func() {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				common.SysError("error closing response body: " + err.Error())
			}
		}()

		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < 6 { // 忽略空行或格式错误的行
				continue
			}
			if data[:6] != "data: " && data[:6] != "[DONE]" {
				continue
			}
			// 重置超时计时器
			ticker.Reset(time.Duration(constant.StreamingTimeout) * time.Second)
			mu.Lock()
			if !strings.HasPrefix(data, "data: [DONE]") {
				streamItems = append(streamItems, data[6:]) // 存储不含前缀的数据
			}
			mu.Unlock()
			dataChan <- data
		}

		// 检查scanner错误
		if err := scanner.Err(); err != nil {
			common.SysError("error reading stream: " + err.Error())
		}
		stopChan <- true
	}()

	service.SetEventStreamHeaders(c)
	isFirst := true
	responseId := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	createdTime := common.GetTimestamp()
	var lastResponseText string
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			if isFirst {
				isFirst = false
				info.SetFirstResponseTime()
			}

			if strings.HasPrefix(data, "data: [DONE]") {
				data = data[:12]
			}

			// 去除可能存在的\r
			data = strings.TrimSuffix(data, "\r")

			if !strings.HasPrefix(data, "data: [DONE]") {
				// 解析响应数据
				data = data[6:] // 移除 "data: " 前缀
				var streamResponse dto.ChatCompletionsStreamResponse
				err := json.Unmarshal([]byte(data), &streamResponse)
				if err != nil {
					common.SysError("error unmarshalling stream response: " + err.Error())
					return true
				}

				// 设置必要的字段
				streamResponse.Id = responseId
				streamResponse.Created = createdTime
				streamResponse.Model = info.UpstreamModelName // 使用实际的模型名称

				// 只在最后一个包含 usage 的响应中处理 token 计算
				if streamResponse.Usage != nil && streamResponse.Usage.TotalTokens != 0 {

					// 计算实际的输入token (缓存命中部分按15%计费)
					usage.PromptTokens = streamResponse.Usage.PromptTokens - int(float64(streamResponse.Usage.PromptCacheHitTokens)*0.85)
					usage.PromptCacheHitTokens = streamResponse.Usage.PromptCacheHitTokens // 确保设置缓存命中token数

					// 计算输出token和总token
					usage.CompletionTokens = streamResponse.Usage.CompletionTokens
					usage.TotalTokens = streamResponse.Usage.PromptTokens + streamResponse.Usage.CompletionTokens

					// 添加详细日志
					//common.LogInfo(c, fmt.Sprintf(
					//	"Token calculation: cacheHit=%d, cacheMiss=%d, 原始Token数= %d,实际计费Token数=%d, completion=%d, total=%d",
					//	cacheHitTokens,
					//	streamResponse.Usage.PromptCacheMissTokens,
					//	streamResponse.Usage.PromptTokens,
					//	usage.PromptTokens,
					//	usage.CompletionTokens,
					//	usage.TotalTokens,
					//))
				}

				// 处理增量响应和工具调用
				if len(streamResponse.Choices) > 0 {
					choice := &streamResponse.Choices[0]
					currentText := choice.Delta.GetContentString()
					choice.Delta.SetContentString(strings.TrimPrefix(currentText, lastResponseText))
					lastResponseText = currentText
					responseTextBuilder.WriteString(choice.Delta.GetContentString())

					// 处理工具调用
					if choice.Delta.ToolCalls != nil {
						if len(choice.Delta.ToolCalls) > toolCount {
							toolCount = len(choice.Delta.ToolCalls)
						}
					}
				}

				responseBytes, err := json.Marshal(streamResponse)
				if err != nil {
					common.SysError("error marshalling stream response: " + err.Error())
					return true
				}
				c.Writer.Write([]byte("data: "))
				c.Writer.Write(responseBytes)
				c.Writer.Write([]byte("\n\n"))
			} else {
				c.Writer.Write([]byte(data))
				c.Writer.Write([]byte("\n\n"))
			}
			return true
		case <-stopChan:
			return false
		case <-ticker.C:
			common.SysError("streaming timeout")
			return false
		}
	})

	return nil, &usage
}

func deepseekHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	var deepseekResp DeepseekTextResponse
	err = json.Unmarshal(responseBody, &deepseekResp)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	// 计算实际的 Token 使用量
	usage := dto.Usage{}
	if deepseekResp.Usage != nil && deepseekResp.Usage.TotalTokens != 0 {

		// 计算实际的输入token (缓存命中部分按15%计费)
		usage.PromptTokens = deepseekResp.Usage.PromptTokens - int(float64(deepseekResp.Usage.PromptCacheHitTokens)*0.85)
		usage.PromptCacheHitTokens = deepseekResp.Usage.PromptCacheHitTokens // 确保设置缓存命中token数
		// 计算输出token和总token
		usage.CompletionTokens = deepseekResp.Usage.CompletionTokens
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	jsonResponse, err := json.Marshal(deepseekResp)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "write_response_body_failed", http.StatusInternalServerError), nil
	}

	return nil, &usage
}
