package ali

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	endpoint    = "business.aliyuncs.com"
	apiVersion  = "2017-12-14"
	apiAction   = "QueryAccountBalance"
	signMethod  = "HMAC-SHA256"
	signVersion = "1.0"
)

// 成功响应结构体
type SuccessResponse struct {
	XMLName   xml.Name     `xml:"QueryAccountBalanceResponse"`
	Code      string       `xml:"Code"`
	Message   string       `xml:"Message"`
	RequestId string       `xml:"RequestId"`
	Success   bool         `xml:"Success"`
	Data      *BalanceData `xml:"Data"`
}

type BalanceData struct {
	AvailableAmount     string `xml:"AvailableAmount"`
	AvailableCashAmount string `xml:"AvailableCashAmount"`
	CreditAmount        string `xml:"CreditAmount"`
	MybankCreditAmount  string `xml:"MybankCreditAmount"`
	Currency            string `xml:"Currency"`
	QuotaLimit          string `xml:"QuotaLimit"`
}

// 错误响应结构体
type ErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	RequestId string   `xml:"RequestId"`
	HostId    string   `xml:"HostId"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Recommend string   `xml:"Recommend"`
}

func RequestQueryAccountBalance(accessKeyId, accessKeySecret string) (*BalanceData, error) {
	// 构造请求参数
	params := url.Values{}
	params.Add("Action", apiAction)
	params.Add("Version", apiVersion)
	params.Add("AccessKeyId", accessKeyId)
	params.Add("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	params.Add("SignatureMethod", signMethod)
	params.Add("SignatureVersion", signVersion)
	params.Add("SignatureNonce", fmt.Sprintf("%d", time.Now().UnixNano()))

	// 生成签名
	signature := createSignature(params, accessKeySecret)
	params.Add("Signature", signature)

	// 发送请求
	client := &http.Client{}
	reqUrl := fmt.Sprintf("https://%s/?%s", endpoint, params.Encode())
	request, _ := http.NewRequest("GET", reqUrl, nil)
	request.Header.Set("Accept", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// 尝试解析为成功响应
	var successResp SuccessResponse
	if err := xml.Unmarshal(body, &successResp); err == nil && successResp.XMLName.Local == "QueryAccountBalanceResponse" {
		return successResp.Data, nil
	}

	// 尝试解析为错误响应
	var errorResp ErrorResponse
	if err := xml.Unmarshal(body, &errorResp); err == nil && errorResp.XMLName.Local == "Error" {
		return nil, errors.New(errorResp.Message)
	}
	return nil, errors.New(fmt.Sprintf("http code: %d", resp.StatusCode))
}

func createSignature(params url.Values, accessKeySecret string) string {
	// 1. 参数排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 构造规范化的查询字符串
	canonicalizedQueryString := ""
	for _, k := range keys {
		keyEscaped := url.QueryEscape(k)
		valueEscaped := url.QueryEscape(params.Get(k))
		canonicalizedQueryString += "&" + keyEscaped + "=" + valueEscaped
	}
	canonicalizedQueryString = strings.TrimPrefix(canonicalizedQueryString, "&")

	// 3. 构造签名字符串
	stringToSign := "GET" + "&" +
		url.QueryEscape("/") + "&" +
		url.QueryEscape(canonicalizedQueryString)

	// 4. 计算HMAC-SHA256
	h := hmac.New(sha256.New, []byte(accessKeySecret+"&"))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}
