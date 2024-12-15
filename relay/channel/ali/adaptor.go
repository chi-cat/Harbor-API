package ali

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel"
	"one-api/relay/channel/openai"
	relaycommon "one-api/relay/common"
	"one-api/relay/constant"
)

type Adaptor struct {
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

// GetRequestURL 根据中继信息生成请求URL。
// 参数:
//
//	info *relaycommon.RelayInfo - 包含中继信息的结构体指针，用于确定请求的模式和基础URL。
//
// 返回值:
//
//	string - 生成的完整请求URL。
//	error - 错误信息，如果出现错误。
func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	// 初始化fullRequestURL变量以存储最终的请求URL。
	var fullRequestURL string

	// 根据info中的RelayMode字段值选择不同的API端点。
	switch info.RelayMode {
	case constant.RelayModeEmbeddings:
		// 在嵌入模式下，构造嵌入API的请求URL。
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/embeddings/text-embedding/text-embedding", info.BaseUrl)
	case constant.RelayModeImagesGenerations:
		// 在图像生成模式下，构造图像合成API的请求URL。
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text2image/image-synthesis", info.BaseUrl)
	default:
		// 对于其他模式，构造兼容模式下的聊天补全API的请求URL。
		fullRequestURL = fmt.Sprintf("%s/compatible-mode/v1/chat/completions", info.BaseUrl)
	}

	// 返回构造的完整请求URL和nil错误，表示操作成功。
	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Set("Authorization", "Bearer "+info.ApiKey)
	if info.IsStream {
		req.Set("X-DashScope-SSE", "enable")
	}
	if c.GetString("plugin") != "" {
		req.Set("X-DashScope-Plugin", c.GetString("plugin"))
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch info.RelayMode {
	case constant.RelayModeEmbeddings:
		baiduEmbeddingRequest := embeddingRequestOpenAI2Ali(*request)
		return baiduEmbeddingRequest, nil
	default:
		aliReq := requestOpenAI2Ali(*request)
		return aliReq, nil
	}
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	aliRequest := oaiImage2Ali(request)
	return aliRequest, nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *dto.OpenAIErrorWithStatusCode) {
	switch info.RelayMode {
	case constant.RelayModeImagesGenerations:
		err, usage = aliImageHandler(c, resp, info)
	case constant.RelayModeEmbeddings:
		err, usage = aliEmbeddingHandler(c, resp)
	default:
		if info.IsStream {
			err, usage = openai.OaiStreamHandler(c, resp, info)
		} else {
			err, usage = openai.OpenaiHandler(c, resp, info.PromptTokens, info.UpstreamModelName)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
