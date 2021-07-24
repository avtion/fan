package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

const (
	accessTokenAPI = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"
	msgAPI         = "https://open.feishu.cn/open-apis/message/v4/send/"
)

var (
	larkDefaultClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: func() http.RoundTripper {
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.MaxIdleConns = 100
			transport.MaxConnsPerHost = 100
			transport.MaxIdleConnsPerHost = 100
			return transport
		}(),
	}

	ErrParamsAreNil  = errors.New("account or msg is nil")
	ErrSendMsgFailed = errors.New("send msg failed")
)

type (
	// Msg 消息抽象
	Msg interface {
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
	dataBytes, _ := jsoniter.Marshal(msg)
	log.Info("", zap.ByteString("raw", dataBytes))
	resp, err := resty.NewWithClient(larkDefaultClient).R().SetContext(ctx).
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
			client:    resty.NewWithClient(larkDefaultClient),
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

func (t *TextMsg) WithOpenID(id string) Msg {
	t.OpenID = id
	return t
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// 卡片消息
//_______________________________________________________________________

type (
	MarkdownMsg struct {
		OpenID  string           `json:"open_id"`
		MsgType string           `json:"msg_type"`
		Content *MarkdownContent `json:"content"`
	}
	MarkdownContent struct {
		Post *MarkdownPost `json:"post"`
	}
	MarkdownPost struct {
		ZhCN *MarkdownZhCN `json:"zh_cn"`
	}
	MarkdownZhCN struct {
		Title   string   `json:"title"`
		Content [][]*Div `json:"content"`
	}
)

var _ Msg = (*MarkdownMsg)(nil)

func NewMarkdownMsg(title string, divs ...[]*Div) *MarkdownMsg {
	msg := &MarkdownMsg{
		MsgType: "post",
		Content: &MarkdownContent{Post: &MarkdownPost{ZhCN: &MarkdownZhCN{
			Title:   title,
			Content: divs,
		}}},
	}
	return msg
}

func (m *MarkdownMsg) WithOpenID(id string) Msg {
	m.OpenID = id
	return m
}

type Div struct {
	Tag    string `json:"tag"`
	Text   string `json:"text,omitempty"`
	Href   string `json:"href,omitempty"`
	UserId string `json:"user_id,omitempty"`
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// 卡片消息
//_______________________________________________________________________

type (
	CardMsg struct {
		OpenID  string `json:"open_id"`
		MsgType string `json:"msg_type"`
		Card    struct {
			Config struct {
				WideScreenMode bool `json:"wide_screen_mode"`
				EnableForward  bool `json:"enable_forward"`
			} `json:"config"`
			Elements []*CardLine `json:"elements"`
			Header   *CardHeader `json:"header"`
		} `json:"card"`
	}

	// CardLine 一行的内容
	CardLine struct {
		Tag      string        `json:"tag"`
		Text     *CardDiv      `json:"text,omitempty"`
		Actions  []*CardAction `json:"actions,omitempty"`
		Elements []*CardDiv    `json:"elements,omitempty"`
	}

	HeaderColor = string
	// CardHeader 卡片标题
	CardHeader struct {
		Title struct {
			Tag     string `json:"tag"`
			Content string `json:"content"`
		} `json:"title"`
		Template HeaderColor `json:"template"`
	}

	ActionType = string
	// CardAction 卡片操作
	CardAction struct {
		Tag   string      `json:"tag"`
		Text  *CardDiv    `json:"text"`
		Url   string      `json:"url"`
		Type  ActionType  `json:"type"`
		Value interface{} `json:"value,omitempty"`
	}

	// CardDiv 内容
	CardDiv struct {
		Content string `json:"content"`
		Tag     string `json:"tag"`
	}
)

const (
	HeaderColorDefault HeaderColor = "wathet" // 消息卡片头部颜色
	HeaderColorSuccess             = "turquoise"
	HeaderColorFailed              = "red"
	actionTypeDefault  ActionType  = "default" // 消息卡片按钮颜色
	actionTypePrimary              = "primary"
	actionTypeDanger               = "danger"
)

var (
	_ Msg = (*CardMsg)(nil)
	_     = []HeaderColor{HeaderColorDefault, HeaderColorSuccess, HeaderColorFailed}
	_     = []ActionType{actionTypeDefault, actionTypePrimary, actionTypeDanger}
)

func (c *CardMsg) WithOpenID(id string) Msg {
	c.OpenID = id
	return c
}

func NewCardMsg(header *CardHeader, lines ...*CardLine) *CardMsg {
	return &CardMsg{
		MsgType: "interactive",
		Card: struct {
			Config struct {
				WideScreenMode bool `json:"wide_screen_mode"`
				EnableForward  bool `json:"enable_forward"`
			} `json:"config"`
			Elements []*CardLine `json:"elements"`
			Header   *CardHeader `json:"header"`
		}{
			Config: struct {
				WideScreenMode bool `json:"wide_screen_mode"`
				EnableForward  bool `json:"enable_forward"`
			}{
				WideScreenMode: true,
				EnableForward:  true,
			},
			Elements: lines,
			Header:   header,
		},
	}
}

// AddContents 新增富文本模块
func (c *CardMsg) AddContents(contents ...string) *CardMsg {
	for _, v := range contents {
		c.Card.Elements = append(c.Card.Elements, &CardLine{
			Tag: "div",
			Text: &CardDiv{
				Content: v,
				Tag:     "lark_md",
			},
		})
	}
	return c
}

// AddNotes 新增备注
func (c *CardMsg) AddNotes(notes ...string) *CardMsg {
	divs := make([]*CardDiv, 0, len(notes))
	for _, note := range notes {
		divs = append(divs, &CardDiv{
			Content: note,
			Tag:     "lark_md",
		})
	}

	// 找一下有没有Note备注模块
	var isAlready bool
	for _, line := range c.Card.Elements {
		if line.Tag == "note" {
			line.Elements = append(line.Elements, divs...)
			isAlready = true
			break
		}
	}
	if isAlready {
		return c
	}

	// 如果没有的话就追加一下
	c.Card.Elements = append(c.Card.Elements, &CardLine{Tag: "note", Elements: divs})
	return c
}

// AddAction 新增操作按钮
func (c *CardMsg) AddAction(actions ...*CardAction) *CardMsg {
	// 找一下有没有action操作模块
	var isAlready bool
	for _, line := range c.Card.Elements {
		if line.Tag == "action" {
			line.Actions = append(line.Actions, actions...)
			isAlready = true
			break
		}
	}
	if isAlready {
		return c
	}

	// 如果没有的话就追加一下
	c.Card.Elements = append(c.Card.Elements, &CardLine{Tag: "action", Actions: actions})
	return c
}

// NewCardHeader 创建消息卡片标题
func NewCardHeader(title string, color HeaderColor) *CardHeader {
	return &CardHeader{
		Title: struct {
			Tag     string `json:"tag"`
			Content string `json:"content"`
		}{
			Tag:     "plain_text",
			Content: title,
		},
		Template: color,
	}
}

// NewCardAction 创建消息卡片操作按钮
func NewCardAction(buttonType ActionType, content string, url string) *CardAction {
	return &CardAction{
		Tag: "button",
		Text: &CardDiv{
			Content: content,
			Tag:     "lark_md",
		},
		Url:  url,
		Type: buttonType,
	}
}
