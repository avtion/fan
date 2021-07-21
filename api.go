package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var apiDefaultClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: func() http.RoundTripper {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.MaxIdleConns = 100
		transport.MaxConnsPerHost = 100
		transport.MaxIdleConnsPerHost = 100
		return transport
	}(),
}

type Api struct {
	ctx    context.Context
	client *resty.Client

	// 参数
	isLogin bool
}

func NewAPI(ctx context.Context, httpClient *http.Client) *Api {
	if httpClient == nil {
		httpClient = apiDefaultClient
	}
	return &Api{
		ctx: ctx,
		client: resty.NewWithClient(httpClient).SetQueryParams(map[string]string{
			"client_id":     "Xqr8w0Uk4ciodqfPwjhav5rdxTaYepD",
			"client_secret": "vD11O6xI9bG3kqYRu9OyPAHkRGxLh4E",
		}),
	}
}

// Login 登录
func (a *Api) Login(username, password string) error {
	// 检查用户名和密码
	if strings.EqualFold(username, "") || strings.EqualFold(password, "") {
		return errors.New("username or password is null")
	}

	// 发送请求
	resp, err := a.client.R().SetContext(a.ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("remember", "true").
		SetBody(func() io.Reader {
			u := make(url.Values, 5)
			u.Set("grant_type", "password")
			u.Set("meican_credential_type", "password")
			u.Set("username", username)
			u.Set("password", password)
			u.Set("username_type", "username")
			return strings.NewReader(u.Encode())
		}()).
		SetResult(new(LoginResp)).
		Post(LoginURL)
	if err != nil {
		return err
	}

	// 判断登录是否成功
	if resp.StatusCode() != http.StatusOK {
		log.Error("login failed", zap.ByteString("元数据", resp.Body()))
		return errors.New("login failed")
	}

	a.isLogin = true
	log.Info("login success", zap.Any("resp", resp.Result()))
	return nil
}

// GetOrders 获取订单
func (a *Api) GetOrders(begin, end time.Time) (*Orders, error) {
	if !a.isLogin {
		return nil, errors.New("not yet login")
	}
	resp, err := a.client.R().SetContext(a.ctx).
		SetQueryParams(map[string]string{
			"withOrderDetail": "false",
			"beginDate":       begin.Format("2006-01-02"),
			"endDate":         end.Format("2006-01-02"),
		}).
		SetResult(new(Orders)).
		Get(OrderURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Error("get orders failed", zap.ByteString("raw", resp.Body()))
		return nil, errors.New("get orders failed")
	}
	return resp.Result().(*Orders), nil
}
