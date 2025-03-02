package tencent

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/service"
	"strconv"
	"strings"
	"time"
)

type QueryTencentBalanceResponseWrapper struct {
	Response QueryTencentBalanceResponse `json:"Response"`
}

type QueryTencentBalanceResponse struct {
	Uin                      int                        `json:"Uin"`
	RealBalance              float64                    `json:"RealBalance"`
	CashAccountBalance       float64                    `json:"CashAccountBalance"`
	IncomeIntoAccountBalance float64                    `json:"IncomeIntoAccountBalance"`
	PresentAccountBalance    float64                    `json:"PresentAccountBalance"`
	FreezeAmount             float64                    `json:"FreezeAmount"`
	OweAmount                float64                    `json:"OweAmount"`
	RequestId                string                     `json:"RequestId"`
	IsAllowArrears           bool                       `json:"IsAllowArrears"`
	IsCreditLimited          bool                       `json:"IsCreditLimited"`
	Balance                  float64                    `json:"Balance"`
	CreditAmount             float64                    `json:"CreditAmount"`
	CreditBalance            float64                    `json:"CreditBalance"`
	RealCreditBalance        float64                    `json:"RealCreditBalance"`
	Error                    *QueryTencentResponseError `json:"Error"`
}

type QueryTencentResponseError struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

func RequestQueryBalanceAcct(ak, sk string) (*QueryTencentBalanceResponse, error) {
	serviceName := "billing"
	host := fmt.Sprintf("%s.tencentcloudapi.com", serviceName)
	version := "2018-07-09"
	action := "DescribeAccountBalance"
	bodyStr := "{}"
	request, _ := http.NewRequest("POST", "https://"+host, bytes.NewBuffer([]byte(bodyStr)))
	timestamp := common.GetTimestamp()
	sign := getTencentSign(timestamp, serviceName, host, action, bodyStr, ak, sk)
	request.Header.Set("Authorization", sign)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Host", host)
	request.Header.Set("X-TC-Action", action)
	request.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	request.Header.Set("X-TC-Version", version)
	resp, err := service.GetHttpClient().Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	result := &QueryTencentBalanceResponseWrapper{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	response := result.Response
	return &response, nil
}

// https://cloud.tencent.com/document/api/555/30995#Golang
func getTencentSign(timestamp int64, service, host, action, body, secId, secKey string) string {
	// build canonical request string

	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		"application/json", host, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	payload := body
	hashedRequestPayload := sha256hex(payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)
	// build string to sign
	algorithm := "TC3-HMAC-SHA256"
	requestTimestamp := strconv.FormatInt(timestamp, 10)
	t := time.Unix(timestamp, 0).UTC()
	// must be the format 2006-01-02, ref to package time for more info
	date := t.Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := sha256hex(canonicalRequest)
	string2sign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		requestTimestamp,
		credentialScope,
		hashedCanonicalRequest)

	// sign string
	secretDate := hmacSha256(date, "TC3"+secKey)
	secretService := hmacSha256(service, secretDate)
	secretKey := hmacSha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(hmacSha256(string2sign, secretKey)))

	// build authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		secId,
		credentialScope,
		signedHeaders,
		signature)
	return authorization
}
