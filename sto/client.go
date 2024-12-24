package sto

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// BaseURL 申通开放平台API地址
	BaseURL = "https://cloudinter-linkgateway.sto.cn/gateway/link.do"

	// DefaultTimeout 默认超时时间
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries 默认最大重试次数
	DefaultMaxRetries = 3
)

// Client 申通开放平台客户端
type Client struct {
	AppKey    string
	AppSecret string
	FromCode  string
	Debug     bool // 是否开启调试模式

	httpClient *http.Client // HTTP客户端
	mu         sync.RWMutex // 保护httpClient

	timeout    time.Duration // 超时时间
	maxRetries int           // 最大重试次数
}

// ClientOption 定义客户端选项
type ClientOption func(*Client)

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithHTTPClient 设置自定义HTTP客户端
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient 创建新的客户端实例
func NewClient(appKey, appSecret, fromCode string, opts ...ClientOption) *Client {
	c := &Client{
		AppKey:     appKey,
		AppSecret:  appSecret,
		FromCode:   fromCode,
		Debug:      false,
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
	}

	// 应用选项
	for _, opt := range opts {
		opt(c)
	}

	// 如果没有提供自定义HTTP客户端，创建默认的
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: c.timeout,
		}
	}

	return c
}

// EnableDebug 开启调试模式
func (c *Client) EnableDebug() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Debug = true
}

// DisableDebug 关闭调试模式
func (c *Client) DisableDebug() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Debug = false
}

// TraceQueryRequest 轨迹查询请求参数
type TraceQueryRequest struct {
	Order         string   `json:"order"`         // 排序方式，asc（升序）或desc（降序）
	WaybillNoList []string `json:"waybillNoList"` // 运单号列表
}

// Validate 验证请求参数
func (r *TraceQueryRequest) Validate() error {
	if len(r.WaybillNoList) == 0 {
		return fmt.Errorf("waybillNoList cannot be empty")
	}
	if r.Order != "" && r.Order != "asc" && r.Order != "desc" {
		return fmt.Errorf("order must be either 'asc' or 'desc'")
	}
	return nil
}

// TraceInfo 物流轨迹信息
type TraceInfo struct {
	WaybillNo         string `json:"waybillNo"`         // 运单号
	OpTime            string `json:"opTime"`            // 操作时间
	OpOrgCode         string `json:"opOrgCode"`         // 操作机构代码
	OpOrgName         string `json:"opOrgName"`         // 操作机构名称
	OpOrgProvinceName string `json:"opOrgProvinceName"` // 操作机构所在省
	OpOrgCityName     string `json:"opOrgCityName"`     // 操作机构所在市
	OpOrgTel          string `json:"opOrgTel"`          // 操作机构电话
	OpEmpCode         string `json:"opEmpCode"`         // 操作员工号
	OpEmpName         string `json:"opEmpName"`         // 操作员姓名
	ScanType          string `json:"scanType"`          // 扫描类型
	Weight            string `json:"weight"`            // 重量
	Memo              string `json:"memo"`              // 备注
	BizEmpCode        string `json:"bizEmpCode"`        // 业务员工号
	BizEmpName        string `json:"bizEmpName"`        // 业务员姓名
	BizEmpPhone       string `json:"bizEmpPhone"`       // 业务员电话
	BizEmpTel         string `json:"bizEmpTel"`         // 业务员固定电话
	NextOrgName       string `json:"nextOrgName"`       // 下一站机构名称
	NextOrgCode       string `json:"nextOrgCode"`       // 下一站机构代码
	IssueName         string `json:"issueName"`         // 问题件名称
	SignoffPeople     string `json:"signoffPeople"`     // 签收人
	ContainerNo       string `json:"containerNo"`       // 集包号
	OrderOrgCode      string `json:"orderOrgCode"`      // 下单机构代码
	OrderOrgName      string `json:"orderOrgName"`      // 下单机构名称
	TransportTaskNo   string `json:"transportTaskNo"`   // 运输任务号
	CarNo             string `json:"carNo"`             // 车牌号
	OpOrgTypeCode     string `json:"opOrgTypeCode"`     // 操作机构类型代码
	PartnerName       string `json:"partnerName"`       // 品牌方名称
}

// TraceQueryResponse 轨迹查询响应
type TraceQueryResponse struct {
	Success   string                 `json:"success"`   // 是否成功
	ErrorCode string                 `json:"errorCode"` // 错误码
	ErrorMsg  string                 `json:"errorMsg"`  // 错误信息
	NeedRetry string                 `json:"needRetry"` // 是否需要重试
	RequestId string                 `json:"requestId"` // 请求ID
	ExpInfo   string                 `json:"expInfo"`   // 异常信息
	Data      map[string][]TraceInfo `json:"data"`      // 运单号对应的轨迹列表
}

// IsSuccess 检查是否成功
func (r *TraceQueryResponse) IsSuccess() bool {
	return r.Success == "true"
}

// ShouldRetry 检查是否需要重试
func (r *TraceQueryResponse) ShouldRetry() bool {
	return r.NeedRetry == "true"
}

// QueryTrace 查询物流轨迹
func (c *Client) QueryTrace(req *TraceQueryRequest) (*TraceQueryResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %v", err)
	}

	// 将请求内容转为JSON
	content, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	// 生成data_digest
	h := md5.New()
	h.Write([]byte(string(content) + c.AppSecret))
	dataDigest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 构建请求参数
	params := url.Values{}
	params.Add("content", string(content))
	params.Add("data_digest", dataDigest)
	params.Add("from_appkey", c.AppKey)
	params.Add("from_code", c.FromCode)
	params.Add("to_appkey", "sto_trace_query")
	params.Add("to_code", "sto_trace_query")
	params.Add("api_name", "STO_TRACE_QUERY_COMMON")

	// 构建完整URL
	requestURL := fmt.Sprintf("%s?%s", BaseURL, params.Encode())

	var resp *TraceQueryResponse
	var lastErr error

	// 重试逻辑
	for i := 0; i <= c.maxRetries; i++ {
		if i > 0 && c.Debug {
			fmt.Printf("Retrying request (attempt %d/%d)\n", i, c.maxRetries)
		}

		resp, lastErr = c.doRequest(requestURL, content, dataDigest)
		if lastErr == nil && !resp.ShouldRetry() {
			break
		}

		if i < c.maxRetries {
			time.Sleep(time.Duration(i+1) * time.Second) // 简单的退避策略
		}
	}

	return resp, lastErr
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(requestURL string, content []byte, dataDigest string) (*TraceQueryResponse, error) {
	if c.Debug {
		fmt.Printf("Request URL: %s\n", requestURL)
		fmt.Printf("Content: %s\n", string(content))
		fmt.Printf("Data Digest: %s\n", dataDigest)
	}

	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	// 发送请求
	c.mu.RLock()
	client := c.httpClient
	debug := c.Debug
	c.mu.RUnlock()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	if debug {
		fmt.Printf("Response Status: %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result TraceQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %v, body: %s", err, string(body))
	}

	return &result, nil
}
