package handlers

import (
	"errors"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	gogpt "github.com/sashabaranov/go-gpt3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xmapst/wechatbot/gpt"
	"github.com/xmapst/wechatbot/service"
	"strings"
)

var _ MessageHandlerInterface = (*UserMessageHandler)(nil)

// UserMessageHandler 私聊消息处理
type UserMessageHandler struct {
	// 接收到消息
	msg *openwechat.Message
	// 发送的用户
	sender *openwechat.User
	// 实现的用户业务
	service service.UserServiceInterface
}

func UserMessageContextHandler() func(ctx *openwechat.MessageContext) {
	return func(ctx *openwechat.MessageContext) {
		msg := ctx.Message
		// 获取用户消息处理器
		handler, err := NewUserMessageHandler(msg)
		if err != nil {
			logrus.Warning(fmt.Sprintf("init user message handler error: %s", err))
			return
		}
		// 处理用户消息
		err = handler.handle()
		if err != nil {
			logrus.Warning(fmt.Sprintf("handle user message error: %s", err))
		}
	}
}

// NewUserMessageHandler 创建私聊处理器
func NewUserMessageHandler(message *openwechat.Message) (MessageHandlerInterface, error) {
	sender, err := message.Sender()
	if err != nil {
		return nil, err
	}
	userService := service.NewUserService(c, sender)
	handler := &UserMessageHandler{
		msg:     message,
		sender:  sender,
		service: userService,
	}

	return handler, nil
}

// handle 处理消息
func (h *UserMessageHandler) handle() error {
	if h.msg.IsText() {
		return h.ReplyText()
	}
	return nil
}

// ReplyText 发送文本消息到群
func (h *UserMessageHandler) ReplyText() error {
	var err error
	// 打印消息内容
	logrus.Info(fmt.Sprintf("Received User %v Text Msg: %v", h.sender.NickName, h.msg.Content))
	// 排除忽略的用户
	if skipUserOrGroup(h.sender) {
		return nil
	}

	// 1.获取上下文，如果字符串为空不处理
	requestText := h.getRequestText()
	if requestText == nil {
		logrus.Info("user message is null")
		return nil
	}
	reply := gpt.Completions(requestText)
	requestText = append(requestText, reply)
	h.service.SetUserSessionContext(requestText)
	// 2.向GPT发起请求，如果回复文本等于空,不回复
	if h.sender.IsSelf() && h.msg.ToUserName == openwechat.FileHelper {
		// 2.1 回复文件助手
		_, err = h.sender.Self().FileHelper().SendText(buildUserReply(reply.Content))
	} else if !h.sender.IsSelf() {
		// 2.2 设置上下文，回复用户
		_, err = h.msg.ReplyText(buildUserReply(reply.Content))
	}
	if err != nil {
		return errors.New(fmt.Sprintf("response user error: %v ", err))
	}

	// 3.返回错误
	return err
}

// getRequestText 获取请求接口的文本，要做一些清晰
func (h *UserMessageHandler) getRequestText() []gogpt.ChatCompletionMessage {
	// 1.去除空格以及换行
	requestText := strings.TrimSpace(h.msg.Content)
	requestText = strings.Trim(h.msg.Content, "\n")

	if len(requestText) >= 2048 {
		requestText = requestText[:2048]
	}
	// 2.检查用户发送文本是否包含结束标点符号
	punctuation := ",.;!?，。！？、…"
	runeRequestText := []rune(requestText)
	lastChar := string(runeRequestText[len(runeRequestText)-1:])
	if strings.Index(punctuation, lastChar) < 0 {
		requestText = requestText + "？" // 判断最后字符是否加了标点，没有的话加上句号，避免openai自动补齐引起混乱。
	}
	// 3.获取上下文，拼接在一起，如果字符长度超出4000，截取为4000。（GPT按字符长度算）
	sessionText := h.service.GetUserSessionContext()
	if sessionText == nil {
		return []gogpt.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "我是一个专注于不标准普通话的小学课本小明",
			},
			{
				Role:    "user",
				Content: requestText,
			},
		}
	}
	sessionText = append(sessionText, gogpt.ChatCompletionMessage{
		Role:    "user",
		Content: requestText,
	})
	// 4.返回请求文本
	return sessionText
}

// buildUserReply 构建用户回复
func buildUserReply(reply string) string {
	// 1.去除空格问号以及换行号，如果为空，返回一个默认值提醒用户
	textSplit := strings.Split(reply, "\n\n")
	if len(textSplit) > 1 {
		trimText := textSplit[0]
		reply = strings.Trim(reply, trimText)
	}
	reply = strings.TrimSpace(reply)

	reply = strings.TrimSpace(reply)
	if reply == "" {
		return "将物质欲望降低些，便没那么多烦恼了"
	}

	// 2.如果用户有配置前缀，加上前缀
	reply = viper.GetString("reply_prefix") + "\n" + reply
	reply = strings.Trim(reply, "\n")

	// 3.返回拼接好的字符串
	return reply
}
