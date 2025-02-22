package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/relay/channel/ali"
	"one-api/relay/channel/volcengine"
	"one-api/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// https://github.com/songquanpeng/one-api/issues/79

type OpenAISubscriptionResponse struct {
	Object             string  `json:"object"`
	HasPaymentMethod   bool    `json:"has_payment_method"`
	SoftLimitUSD       float64 `json:"soft_limit_usd"`
	HardLimitUSD       float64 `json:"hard_limit_usd"`
	SystemHardLimitUSD float64 `json:"system_hard_limit_usd"`
	AccessUntil        int64   `json:"access_until"`
}

type OpenAIUsageDailyCost struct {
	Timestamp float64 `json:"timestamp"`
	LineItems []struct {
		Name string  `json:"name"`
		Cost float64 `json:"cost"`
	}
}

type OpenAICreditGrants struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalAvailable float64 `json:"total_available"`
}

type OpenAIUsageResponse struct {
	Object string `json:"object"`
	//DailyCosts []OpenAIUsageDailyCost `json:"daily_costs"`
	TotalUsage float64 `json:"total_usage"` // unit: 0.01 dollar
}

type OpenAISBUsageResponse struct {
	Msg  string `json:"msg"`
	Data *struct {
		Credit string `json:"credit"`
	} `json:"data"`
}

type AIProxyUserOverviewResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode int    `json:"error_code"`
	Data      struct {
		TotalPoints float64 `json:"totalPoints"`
	} `json:"data"`
}

type API2GPTUsageResponse struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalRemaining float64 `json:"total_remaining"`
}

type APGC2DGPTUsageResponse struct {
	//Grants         interface{} `json:"grants"`
	Object         string  `json:"object"`
	TotalAvailable float64 `json:"total_available"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
}

type DeepSeekUserBalanceResponse struct {
	IsAvailable  bool                  `json:"is_available"`
	BalanceInfos []DeepSeekUserBalance `json:"balance_infos"`
}

type DeepSeekCurrencyEnum string

const (
	CNY DeepSeekCurrencyEnum = "CNY"
	USD DeepSeekCurrencyEnum = "USD"
)

type DeepSeekUserBalance struct {
	Currency        DeepSeekCurrencyEnum `json:"currency"`
	TotalBalance    string               `json:"total_balance"`
	GrantedBalance  string               `json:"granted_balance"`
	ToppedUpBalance string               `json:"topped_up_balance"`
}

type SiliconflowBalanceResponse struct {
	Code int                    `json:"code"`
	Data SiliconflowBalanceData `json:"data"`
}

type SiliconflowBalanceData struct {
	TotalBalance string `json:"totalBalance"`
}

// GetAuthHeader get auth header
func GetAuthHeader(token string) http.Header {
	h := http.Header{}
	h.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return h
}

func GetResponseBody(method, url string, channel *model.Channel, headers http.Header) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	for k := range headers {
		req.Header.Add(k, headers.Get(k))
	}
	res, err := service.GetHttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func updateChannelCloseAIBalance(channel *model.Channel) (float64, error) {
	url := fmt.Sprintf("%s/dashboard/billing/credit_grants", channel.GetBaseURL())
	body, err := GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))

	if err != nil {
		return 0, err
	}
	response := OpenAICreditGrants{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(response.TotalAvailable)
	return response.TotalAvailable, nil
}

func updateChannelOpenAISBBalance(channel *model.Channel) (float64, error) {
	url := fmt.Sprintf("https://api.openai-sb.com/sb-api/user/status?api_key=%s", channel.Key)
	body, err := GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := OpenAISBUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if response.Data == nil {
		return 0, errors.New(response.Msg)
	}
	balance, err := strconv.ParseFloat(response.Data.Credit, 64)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(balance)
	return balance, nil
}

func updateChannelAIProxyBalance(channel *model.Channel) (float64, error) {
	url := "https://aiproxy.io/api/report/getUserOverview"
	headers := http.Header{}
	headers.Add("Api-Key", channel.Key)
	body, err := GetResponseBody("GET", url, channel, headers)
	if err != nil {
		return 0, err
	}
	response := AIProxyUserOverviewResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if !response.Success {
		return 0, fmt.Errorf("code: %d, message: %s", response.ErrorCode, response.Message)
	}
	channel.UpdateBalance(response.Data.TotalPoints)
	return response.Data.TotalPoints, nil
}

func updateChannelAPI2GPTBalance(channel *model.Channel) (float64, error) {
	url := "https://api.api2gpt.com/dashboard/billing/credit_grants"
	body, err := GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))

	if err != nil {
		return 0, err
	}
	response := API2GPTUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(response.TotalRemaining)
	return response.TotalRemaining, nil
}

func updateChannelAIGC2DBalance(channel *model.Channel) (float64, error) {
	url := "https://api.aigc2d.com/dashboard/billing/credit_grants"
	body, err := GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := APGC2DGPTUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(response.TotalAvailable)
	return response.TotalAvailable, nil
}

func updateChannelSiliconflowBalance(channel *model.Channel) (float64, error) {
	url := fmt.Sprintf("%s/v1/user/info", *channel.BaseURL)
	body, err := GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := &SiliconflowBalanceResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return 0, err
	}
	if response.Code != 20000 {
		return 0, err
	}
	balanceOfInfo, err := strconv.ParseFloat(response.Data.TotalBalance, 64)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(balanceOfInfo / common.USD2RMB)
	return balanceOfInfo / common.USD2RMB, nil
}

func updateChannelDeepSeekBalance(channel *model.Channel) (float64, error) {
	url := fmt.Sprintf("%s/user/balance", *channel.BaseURL)
	body, err := GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := &DeepSeekUserBalanceResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return 0, err
	}
	if !response.IsAvailable {
		return 0, err
	}
	var balance = 0.0
	for _, info := range response.BalanceInfos {
		balanceOfInfo, err := strconv.ParseFloat(info.TotalBalance, 64)
		if err != nil {
			return 0, err
		}
		switch info.Currency {
		case CNY:
			balance += balanceOfInfo / common.USD2RMB
		case USD:
			balance += balanceOfInfo
		}
	}
	channel.UpdateBalance(balance)
	return balance, nil
}

type AccessKeys struct {
	AccessKey       string `json:"access_key"`
	AccessKeySecret string `json:"access_key_secret"`
}

func updateChannelVolcengineBalance(channel *model.Channel) (float64, error) {
	sensitiveInfo := channel.OtherSensitiveInfo
	if sensitiveInfo == nil || *sensitiveInfo == "" {
		return 0, errors.New("没有配置火山平台的ak或sk")
	}
	credentials := &AccessKeys{}
	err := json.Unmarshal([]byte(*sensitiveInfo), credentials)
	if err != nil {
		return 0, err
	}
	acct, err := volcengine.RequestQueryBalanceAcct(credentials.AccessKey, credentials.AccessKeySecret)
	if err != nil {
		return 0, err
	}
	if acct.Error != nil {
		return 0, errors.New(fmt.Sprintf("%v", *acct.Error))
	}
	balanceOfInfo, err := strconv.ParseFloat(*acct.Result.AvailableBalance, 64)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(balanceOfInfo / common.USD2RMB)
	return balanceOfInfo / common.USD2RMB, nil
}

func updateChannelAliBalance(channel *model.Channel) (float64, error) {
	sensitiveInfo := channel.OtherSensitiveInfo
	if sensitiveInfo == nil || *sensitiveInfo == "" {
		return 0, errors.New("没有配置阿里平台的ak或sk")
	}
	credentials := &AccessKeys{}
	err := json.Unmarshal([]byte(*sensitiveInfo), credentials)
	if err != nil {
		return 0, err
	}
	acct, err := ali.RequestQueryAccountBalance(credentials.AccessKey, credentials.AccessKeySecret)
	if err != nil {
		return 0, err
	}
	balance, err := strconv.ParseFloat(acct.AvailableAmount, 64)
	if err != nil {
		return 0, err
	}
	if acct.Currency == "CNY" {
		balance = balance / common.USD2RMB
	}
	channel.UpdateBalance(balance)
	return balance, nil
}

func updateChannelBalance(channel *model.Channel) (float64, error) {
	baseURL := common.ChannelBaseURLs[channel.Type]
	if channel.GetBaseURL() == "" {
		channel.BaseURL = &baseURL
	}
	switch channel.Type {
	case common.ChannelTypeOpenAI:
		if channel.GetBaseURL() != "" {
			baseURL = channel.GetBaseURL()
		}
	case common.ChannelTypeAzure:
		return 0, errors.New("尚未实现")
	case common.ChannelTypeCustom:
		baseURL = channel.GetBaseURL()
	//case common.ChannelTypeOpenAISB:
	//	return updateChannelOpenAISBBalance(channel)
	case common.ChannelTypeAIProxy:
		return updateChannelAIProxyBalance(channel)
	case common.ChannelTypeAPI2GPT:
		return updateChannelAPI2GPTBalance(channel)
	case common.ChannelTypeAIGC2D:
		return updateChannelAIGC2DBalance(channel)
	case common.ChannelTypeDeepseek:
		return updateChannelDeepSeekBalance(channel)
	case common.ChannelTypeSiliconFlow:
		return updateChannelSiliconflowBalance(channel)
	case common.ChannelTypeVolcEngine:
		return updateChannelVolcengineBalance(channel)
	case common.ChannelTypeAli:
		return updateChannelAliBalance(channel)
	default:
		return 0, errors.New("尚未实现")
	}
	url := fmt.Sprintf("%s/v1/dashboard/billing/subscription", baseURL)

	body, err := GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	subscription := OpenAISubscriptionResponse{}
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	startDate := fmt.Sprintf("%s-01", now.Format("2006-01"))
	endDate := now.Format("2006-01-02")
	if !subscription.HasPaymentMethod {
		startDate = now.AddDate(0, 0, -100).Format("2006-01-02")
	}
	url = fmt.Sprintf("%s/v1/dashboard/billing/usage?start_date=%s&end_date=%s", baseURL, startDate, endDate)
	body, err = GetResponseBody("GET", url, channel, GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	usage := OpenAIUsageResponse{}
	err = json.Unmarshal(body, &usage)
	if err != nil {
		return 0, err
	}
	balance := subscription.HardLimitUSD - usage.TotalUsage/100
	channel.UpdateBalance(balance)
	return balance, nil
}

func UpdateChannelBalance(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	balance, err := updateChannelBalance(channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"balance": balance,
	})
	return
}

func updateAllChannelsBalance() error {
	channels, err := model.GetAllChannels(0, 0, true, false)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if channel.Status != common.ChannelStatusEnabled {
			continue
		}
		// TODO: support Azure
		if channel.Type != common.ChannelTypeOpenAI && channel.Type != common.ChannelTypeCustom {
			continue
		}
		balance, err := updateChannelBalance(channel)
		if err != nil {
			continue
		} else {
			// err is nil & balance <= 0 means quota is used up
			if balance <= 0 {
				service.DisableChannel(channel.Id, channel.Name, "余额不足")
			}
		}
		time.Sleep(common.RequestInterval)
	}
	return nil
}

func UpdateAllChannelsBalance(c *gin.Context) {
	// TODO: make it async
	err := updateAllChannelsBalance()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func AutomaticallyUpdateChannels(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		common.SysLog("updating all channels")
		_ = updateAllChannelsBalance()
		common.SysLog("channels update done")
	}
}
