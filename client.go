package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	uuid "github.com/satori/go.uuid"
)

type Client struct {
	accept              string
	contentType         string
	date                string
	appKey              string
	appSecret           string
	xCaNonce            string
	xCaSignatureHeaders string
	xCaSignature        string
	xCaTimestamp        int64
	xCaSignatureMethod  SignatureMethod
	ctx                 context.Context
	body                any               // post请求传入
	result              any               // result
	formData            map[string]string // formdata
	queryParams         map[string]string // 参数
	resty               *resty.Client
}

// NewClient 初始化连接
// 默认使用HmacSHA256加密方式
func NewClient(appKey, appSecret string) *Client {
	return &Client{
		accept:              "application/json; charset=utf-8",
		appKey:              appKey,
		appSecret:           appSecret,
		xCaSignatureMethod:  HmacSHA256,
		xCaSignatureHeaders: "x-ca-timestamp,x-ca-key,x-ca-nonce,x-ca-signature-method",
		resty:               resty.New(),
		formData:            map[string]string{},
		queryParams:         map[string]string{},
	}
}

// SetResty 请求设置
func (c *Client) SetResty(resty *resty.Client) *Client {
	c.resty = resty
	return c
}

// SetContext 设置上下文
func (c *Client) SetContext(ctx context.Context) *Client {
	c.ctx = ctx
	return c
}

// contex 上下文
func (c *Client) contex() context.Context {
	if c.ctx == nil {
		return context.Background()
	}

	return c.ctx
}

// SetXCaSignatureMethod 设置加密方式
func (c *Client) SetXCaSignatureMethod(xCaSignatureMethod SignatureMethod) *Client {
	c.xCaSignatureMethod = xCaSignatureMethod
	return c
}

// SetqueryParams 批量设置参数
func (c *Client) SetqueryParams(params map[string]string) *Client {
	for p, v := range params {
		c.queryParams[p] = v
	}

	return c
}

// SetqueryParam 设置单个参数
func (c *Client) SetQueryParam(param, value string) *Client {
	c.queryParams[param] = value
	return c
}

// SetBody 设置body参数
// 支持map、struct、string、[]byte
// 具体可查看resty的文档:https://pkg.go.dev/github.com/go-resty/resty/v2@v2.7.0#Request.SetBody
func (c *Client) SetBody(body any) *Client {
	c.body = body
	return c
}

// SetFormData 设置formData参数
func (c *Client) SetFormData(formData map[string]string) *Client {
	c.formData = formData
	return c
}

// SetResult 将结果解析到传入的参数
func (c *Client) SetResult(result any) *Client {
	c.result = getPointer(result)
	return c
}

func getPointer(v interface{}) interface{} {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		return v
	}
	return reflect.New(vv.Type()).Interface()
}

// genUUID 生成uuid
func (c *Client) genUUID() string {
	return uuid.NewV4().String()
}

// genTime 生成时间
func (c *Client) genTime() (string, int64) {
	now := time.Now()
	return now.Format(`Mon, 02 Jan 2006 15:04:05 GMT-07:00`), now.UnixMilli()
}

// getURLPath 优先取fromdata,queryParams,如果没有就取url里面的
func (c *Client) getURLPath(urlStr string, httpMethod string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	path := parsedURL.Path
	if httpMethod == http.MethodGet {
		if len(c.queryParams) == 0 {
			queryParams := parsedURL.Query()
			if len(queryParams) > 0 {
				return strings.TrimPrefix(urlStr, parsedURL.Scheme+"://"+parsedURL.Host), nil
			}

			return path, nil
		}
	}

	var params string
	if httpMethod != http.MethodGet {
		if len(c.formData) != 0 {
			for p, v := range c.formData {
				_params := p + "=" + v
				params += _params
			}
		}
	}

	if len(c.queryParams) != 0 {
		for p, v := range c.queryParams {
			_params := p + "=" + v
			params += _params
		}
	}

	if len(c.queryParams) == 0 && len(c.formData) == 0 {
		return path, nil
	}

	path += "?" + params
	return path, nil
}

// setDefaultMultiField 设置多个默认字段
func (c *Client) setDefaultMultiField(url, httpMethod string) error {
	c.date, c.xCaTimestamp = c.genTime()
	c.xCaNonce = c.genUUID()

	path, err := c.getURLPath(url, httpMethod)
	if err != nil {
		return err
	}

	signature := fmt.Sprintf(signatureStr,
		httpMethod,
		c.accept,
		c.contentType,
		c.date,
		c.appKey,
		c.xCaNonce,
		c.xCaSignatureMethod,
		c.xCaTimestamp,
		path,
	)

	fmt.Println(signature)

	c.xCaSignature, err = c.genEncryptStr(signature)
	if err != nil {
		return err
	}

	return nil
}

// GET 发起GET请求
func (c *Client) GET(url string) (*resty.Response, error) {
	c.contentType = "application/x-www-form-urlencoded; charset=utf-8"
	if err := c.setDefaultMultiField(url, http.MethodGet); err != nil {
		return nil, err
	}

	resty := c.resty.R().SetContext(c.contex())
	resty.SetHeaders(map[string]string{
		"accept":                 c.accept,
		"content-type":           c.contentType,
		"date":                   c.date,
		"x-ca-key":               c.appKey,
		"x-ca-nonce":             c.xCaNonce,
		"x-ca-signature-method":  string(c.xCaSignatureMethod),
		"x-ca-timestamp":         strconv.Itoa(int(c.xCaTimestamp)),
		"x-ca-signature-headers": c.xCaSignatureHeaders,
		"x-ca-signature":         c.xCaSignature,
	})

	if c.result != nil {
		resty = resty.SetResult(c.result)
	}

	if len(c.queryParams) > 0 {
		resty.SetQueryParams(c.queryParams)
	}

	if len(c.formData) > 0 {
		resty.SetFormData(c.formData)
	}

	return resty.Get(url)
}

// notGet 非GET请求
func (c *Client) notGET(url, httpMethod string) (*resty.Request, error) {
	c.contentType = "application/json; charset=utf-8"
	if err := c.setDefaultMultiField(url, httpMethod); err != nil {
		return nil, err
	}

	resty := c.resty.R().SetContext(c.contex())
	resty.SetHeaders(map[string]string{
		"accept":                 c.accept,
		"content-type":           c.contentType,
		"date":                   c.date,
		"x-ca-key":               c.appKey,
		"x-ca-nonce":             c.xCaNonce,
		"x-ca-signature-method":  string(c.xCaSignatureMethod),
		"x-ca-timestamp":         strconv.Itoa(int(c.xCaTimestamp)),
		"x-ca-signature-headers": c.xCaSignatureHeaders,
		"x-ca-signature":         c.xCaSignature,
	})
	if c.result != nil {
		resty = resty.SetResult(c.result)
	}

	if len(c.queryParams) > 0 {
		resty.SetQueryParams(c.queryParams)
	}

	if len(c.formData) > 0 {
		resty.SetFormData(c.formData)
	}

	if c.body != nil {
		resty.SetBody(c.body)
	}

	return resty, nil
}

// POST 发起POST请求
func (c *Client) POST(url string) (*resty.Response, error) {
	resty, err := c.notGET(url, http.MethodPost)
	if err != nil {
		return nil, err
	}
	return resty.Post(url)
}

// DELETE 发起DELETE请求
func (c *Client) DELETE(url string) (*resty.Response, error) {
	resty, err := c.notGET(url, http.MethodDelete)
	if err != nil {
		return nil, err
	}
	return resty.Delete(url)
}

// PUT 发起PUT请求
func (c *Client) PUT(url string) (*resty.Response, error) {
	resty, err := c.notGET(url, http.MethodPut)
	if err != nil {
		return nil, err
	}
	return resty.Put(url)
}

// genEncryptStr 生成加密字符
func (c *Client) genEncryptStr(str string) (string, error) {
	if c.xCaSignatureMethod == HmacSHA256 {
		return hmacSHA256Encrypt(c.appSecret, str)
	}
	return hmacSHA1Encrypt(c.appSecret, str)
}
