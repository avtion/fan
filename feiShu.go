package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type (
	// Msg 消息抽象
	Msg interface {
		GetReader() io.Reader
		WithOpenID(id string) Msg
	}
	// MsgResp 响应结果
	MsgResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			MessageId string `json:"message_id"`
		} `json:"data"`
	}
)

const (
	accessTokenAPI = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"
	msgAPI         = "https://open.feishu.cn/open-apis/message/v4/send/"
)

var (
	defaultClient = &http.Client{}

	ErrParamsAreNil  = errors.New("account or msg is nil")
	ErrSendMsgFailed = errors.New("send msg failed")
)

func SendMsg(ctx context.Context, account *Account, msg Msg) error {
	if account == nil || msg == nil {
		return ErrParamsAreNil
	}

	// 如果该用户使用自建应用机器人
	if _, isExist := globalCfg.FeiShu[account.FeiShuRobot]; isExist {
		if err := SendMsgToRobot(ctx, account, msg); err == nil {
			return nil
		}
		// 降级处理 - 使用webhook发送通知
	}

	return SendMsgToWebHook(ctx, account, msg)
}

func SendMsgToRobot(ctx context.Context, ac *Account, msg Msg) error {
	if ac.FeiShuRobot == "" {
		return errors.New("account no robot")
	}
	r, isExist := robotMapping[ac.FeiShuRobot]
	if !isExist {
		return errors.New("system no robot")
	}
	resp, err := r.client.R().SetContext(ctx).
		SetAuthToken(r.accessToken).
		SetBody(msg.WithOpenID(ac.OpenID)).
		SetResult(new(MsgResp)).
		Post(msgAPI)
	if err != nil {
		log.Error(
			"send msg to robot failed",
			zap.String("robotName", ac.FeiShuRobot),
			zap.Error(err),
		)
		return err
	}

	if realResp := resp.Result().(*MsgResp); realResp.Code != 0 {
		log.Error(
			"send msg to robot failed",
			zap.Any("resp", realResp),
		)
		return ErrSendMsgFailed
	}
	return nil
}

func SendMsgToWebHook(ctx context.Context, ac *Account, msg Msg) error {
	resp, err := resty.NewWithClient(defaultClient).SetTimeout(10 * time.Second).R().SetContext(ctx).
		SetBody(msg).
		SetResult(new(MsgResp)).
		Post(ac.FeiShuWebHook)
	if err != nil {
		log.Error(
			"send msg to webhook failed",
			zap.String("webHook", ac.FeiShuWebHook),
			zap.Error(err),
		)
		return err
	}
	if realResp := resp.Result().(*MsgResp); realResp.Code != 0 {
		log.Error(
			"send msg to webhook failed",
			zap.String("webHook", ac.FeiShuWebHook),
			zap.Any("resp", realResp),
		)
		return ErrSendMsgFailed
	}
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Robot企业自建应用机器人
//_______________________________________________________________________

type Robot struct {
	client                        *resty.Client
	appID, appSecret, accessToken string
}

var robotMapping = make(map[string]*Robot, 0)

func InitRobots(c *cfg) {
	robotMapping = make(map[string]*Robot, len(c.FeiShu))
	for k, v := range c.FeiShu {
		if v.AppID == "" || v.AppSecret == "" {
			continue
		}
		r := &Robot{
			client:    resty.NewWithClient(defaultClient),
			appID:     v.AppID,
			appSecret: v.AppSecret,
		}
		if err := r.getAccessToken(context.Background()); err == nil {
			robotMapping[k] = r
		}
	}
	log.Info("init robot successfully")
}

type (
	accessTokenReq struct {
		AppId     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}
	accessTokenResp struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
)

// 获取鉴权密钥
func (a *Robot) getAccessToken(ctx context.Context) error {
	respData := new(accessTokenResp)
	resp, err := a.client.R().SetContext(ctx).
		SetBody(&accessTokenReq{
			AppId:     a.appID,
			AppSecret: a.appSecret,
		}).
		SetResult(respData).
		Post(accessTokenAPI)
	if err != nil ||
		resp.StatusCode() != http.StatusOK ||
		respData.Code != 0 {
		log.Error("get access token failed",
			zap.Any("resp", respData),
			zap.Error(err),
		)
		return err
	}

	// 设置accessToken
	a.accessToken = respData.TenantAccessToken
	log.Debug("get access token successfully")
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// 文本消息
//_______________________________________________________________________

type TextMsg struct {
	OpenID  string `json:"open_id"`
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

var _ Msg = (*TextMsg)(nil)

func NewTextMsg(content string) *TextMsg {
	return &TextMsg{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: content,
		},
	}
}

func (t *TextMsg) GetReader() io.Reader {
	data, _ := jsoniter.Marshal(t)
	return bytes.NewReader(data)
}

func (t *TextMsg) WithOpenID(id string) Msg {
	t.OpenID = id
	return t
}
