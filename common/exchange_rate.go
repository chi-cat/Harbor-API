package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func GetString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

var (
	exchangeRate    float64 = 7.3 // 默认汇率
	exchangeRateMux sync.RWMutex
)

type ExchangeRateResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Rate float64 `json:"rate"`
	} `json:"result"`
}

// UpdateExchangeRate 更新美元对人民币汇率
// 使用 API 获取实时汇率数据
func UpdateExchangeRate() error {
	// 这里使用 API-Ninjas 的汇率 API，你需要在环境变量中设置 EXCHANGE_RATE_API_KEY
	url := "https://api.api-ninjas.com/v1/exchangerate?pair=USD_CNY"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 从环境变量获取 API key
	apiKey := GetString("EXCHANGE_RATE_API_KEY", "")
	if apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result ExchangeRateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("API 返回失败")
	}

	// 更新汇率
	exchangeRateMux.Lock()
	exchangeRate = result.Result.Rate
	exchangeRateMux.Unlock()

	return nil
}

// GetExchangeRate 获取当前汇率
func GetExchangeRate() float64 {
	exchangeRateMux.RLock()
	defer exchangeRateMux.RUnlock()
	return exchangeRate
}

// StartExchangeRateUpdateTask 启动汇率更新任务
func StartExchangeRateUpdateTask() {
	// 每6小时更新一次汇率
	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		for range ticker.C {
			if err := UpdateExchangeRate(); err != nil {
				fmt.Printf("更新汇率失败: %v\n", err)
			}
		}
	}()
}
