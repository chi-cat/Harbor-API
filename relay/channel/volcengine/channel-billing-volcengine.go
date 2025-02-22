package volcengine

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"one-api/service"
	"strings"
	"time"
)

const Version = "2022-01-01"
const Service = "billing"
const Region = "cn-beijing"
const Host = "open.volcengineapi.com"

// 身份证明结构体
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Service         string
	Region          string
}

// 第二步：准备需要用到的结构体定义。
// 签算请求结构体
type RequestParam struct {
	Body      []byte
	Method    string
	Date      time.Time
	Path      string
	Host      string
	QueryList url.Values
}

// 签算结果结构体
type SignRequest struct {
	XDate          string
	Host           string
	ContentType    string
	XContentSha256 string
	Authorization  string
}

// 第一步：准备辅助函数。
// sha256非对称加密
func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

// sha256 hash算法
func hashSHA256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

type QueryBalanceAcctResp struct {
	Error  *map[string]any `json:"Error"`
	Result *QueryBalanceAcctOutput
}

type QueryBalanceAcctOutput struct {
	_ struct{} `type:"structure" json:",omitempty"`

	AccountID *int32 `type:"int32" json:",omitempty"`

	ArrearsBalance *string `type:"string" json:",omitempty"`

	AvailableBalance *string `type:"string" json:",omitempty"`

	CashBalance *string `type:"string" json:",omitempty"`

	CreditLimit *string `type:"string" json:",omitempty"`

	FreezeAmount *string `type:"string" json:",omitempty"`
}

func RequestQueryBalanceAcct(ak, sk string) (*QueryBalanceAcctResp, error) {
	// 创建 HTTP 请求实例。该实例会在后续用到。
	request, _ := http.NewRequest("GET", "https://"+Host+"/", nil)
	urlVales := url.Values{}
	urlVales["Action"] = []string{"QueryBalanceAcct"}
	urlVales["Version"] = []string{Version}
	request.URL.RawQuery = urlVales.Encode()

	// 第五步：创建身份证明。其中的 Service 和 Region 字段是固定的。ak 和 sk 分别代表 AccessKeyID 和 SecretAccessKey。同时需要初始化签名结构体。一些签名计算时需要的属性也在这里处理。
	// 初始化身份证明
	credential := Credentials{
		AccessKeyID:     ak,
		SecretAccessKey: sk,
		Service:         Service,
		Region:          Region,
	}
	// 初始化签名结构体
	requestParam := RequestParam{
		Body:      nil,
		Host:      request.Host,
		Path:      "/",
		Method:    request.Method,
		Date:      time.Now().UTC(),
		QueryList: request.URL.Query(),
	}
	// 第六步：接下来开始计算签名。在计算签名前，先准备好用于接收签算结果的 signResult 变量，并设置一些参数。
	// 初始化签名结果的结构体
	xDate := requestParam.Date.Format("20060102T150405Z")
	shortXDate := xDate[:8]
	XContentSha256 := hashSHA256(requestParam.Body)
	contentType := "application/json"
	signResult := SignRequest{
		Host:           requestParam.Host, // 设置Host
		XContentSha256: XContentSha256,    // 加密body
		XDate:          xDate,             // 设置标准化时间
		ContentType:    contentType,       // 设置Content-Type 为 application/json
	}
	// 第七步：计算 Signature 签名。
	signedHeadersStr := strings.Join([]string{"content-type", "host", "x-content-sha256", "x-date"}, ";")
	canonicalRequestStr := strings.Join([]string{
		requestParam.Method,
		requestParam.Path,
		request.URL.RawQuery,
		strings.Join([]string{"content-type:" + contentType, "host:" + requestParam.Host, "x-content-sha256:" + XContentSha256, "x-date:" + xDate}, "\n"),
		"",
		signedHeadersStr,
		XContentSha256,
	}, "\n")
	hashedCanonicalRequest := hashSHA256([]byte(canonicalRequestStr))
	credentialScope := strings.Join([]string{shortXDate, credential.Region, credential.Service, "request"}, "/")
	stringToSign := strings.Join([]string{
		"HMAC-SHA256",
		xDate,
		credentialScope,
		hashedCanonicalRequest,
	}, "\n")
	kDate := hmacSHA256([]byte(credential.SecretAccessKey), shortXDate)
	kRegion := hmacSHA256(kDate, credential.Region)
	kService := hmacSHA256(kRegion, credential.Service)
	kSigning := hmacSHA256(kService, "request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))
	signResult.Authorization = fmt.Sprintf("HMAC-SHA256 Credential=%s, SignedHeaders=%s, Signature=%s", credential.AccessKeyID+"/"+credentialScope, signedHeadersStr, signature)
	// 第八步：将 Signature 签名写入HTTP Header 中，并发送 HTTP 请求。
	// 设置经过签名的5个HTTP Header
	request.Header.Set("Host", signResult.Host)
	request.Header.Set("Content-Type", signResult.ContentType)
	request.Header.Set("X-Date", signResult.XDate)
	request.Header.Set("X-Content-Sha256", signResult.XContentSha256)
	request.Header.Set("Authorization", signResult.Authorization)
	resp, err := service.GetHttpClient().Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	result := &QueryBalanceAcctResp{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
