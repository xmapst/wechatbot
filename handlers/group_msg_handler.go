package handlers

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	gogpt "github.com/sashabaranov/go-gpt3"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/wechatbot/gpt"
	"github.com/xmapst/wechatbot/service"
	"strings"
)

var _ MessageHandlerInterface = (*GroupMessageHandler)(nil)

// GroupMessageHandler 群消息处理
type GroupMessageHandler struct {
	// 获取自己
	self *openwechat.Self
	// 群
	group *openwechat.Group
	// 接收到消息
	msg *openwechat.Message
	// 发送的用户
	sender *openwechat.User
	// 实现的用户业务
	service service.UserServiceInterface
}

func GroupMessageContextHandler() func(ctx *openwechat.MessageContext) {
	return func(ctx *openwechat.MessageContext) {
		msg := ctx.Message
		// 获取群组消息处理器
		handler, err := NewGroupMessageHandler(msg)
		if err != nil {
			logrus.Warningf("init group message handler error: %s", err)
			return
		}
		// 处理用户消息
		err = handler.handle()
		if err != nil {
			logrus.Warningf("handle group message error: %s", err)
		}
	}
}

// NewGroupMessageHandler 创建群消息处理器
func NewGroupMessageHandler(message *openwechat.Message) (MessageHandlerInterface, error) {
	sender, err := message.Sender()
	if err != nil {
		return nil, err
	}

	group := &openwechat.Group{User: sender}
	groupSender, err := message.SenderInGroup()
	if err != nil {
		return nil, err
	}
	userService := service.NewUserService(c, groupSender)
	handler := &GroupMessageHandler{
		self:    sender.Self(),
		msg:     message,
		group:   group,
		sender:  groupSender,
		service: userService,
	}
	return handler, nil

}

// handle 处理消息
func (g *GroupMessageHandler) handle() error {
	if g.msg.IsText() {
		return g.ReplyText()
	}
	return nil
}

// ReplyText 发息送文本消到群
func (g *GroupMessageHandler) ReplyText() error {
	var err error
	// 打印消息内容
	logrus.Infof("Received User %v From Group %v Text Msg: %v", g.sender.NickName, g.group.NickName, g.msg.Content)
	// 排除忽略的群组
	if skipUserOrGroup(g.group.User) || skipUserOrGroup(g.sender) {
		return nil
	}

	// 1.不是@的不处理
	if !g.msg.IsAt() || g.sender.IsSelf() {
		return nil
	}

	// 2.获取请求的文本，如果为空字符串不处理
	requestText := g.getRequestText()
	if requestText == nil {
		logrus.Info("user message is null")
		return nil
	}

	// 3.请求GPT获取回复
	reply := gpt.Completions(requestText)
    logrus.Infof("Reply User %v From Group %v Text Msg: %v", g.sender.NickName, g.group.NickName, reply.Content)
	// 4.设置上下文，并响应信息给用户
	requestText = append(requestText, reply)
	g.service.SetUserSessionContext(requestText)
	_, err = g.msg.ReplyText(g.buildReplyText(reply.Content))
	if err != nil {
		return fmt.Errorf("response user error: %v ", err)
	}

	// 5.返回错误信息
	return err
}

// getRequestText 获取请求接口的文本，要做一些清洗
func (g *GroupMessageHandler) getRequestText() []gogpt.ChatCompletionMessage {
	// 1.去除空格以及换行
	requestText := strings.TrimSpace(g.msg.Content)
	requestText = strings.Trim(g.msg.Content, "\n")

	// 2.替换掉当前用户名称
	replaceText := "@" + g.self.NickName
	requestText = strings.TrimSpace(strings.ReplaceAll(g.msg.Content, replaceText, ""))
	if requestText == "" {
		return nil
	}
	if len(requestText) >= 2048 {
		requestText = requestText[:2048]
	}
	// 3.检查用户发送文本是否包含结束标点符号
	punctuation := ",.;!?，。！？、…"
	runeRequestText := []rune(requestText)
	lastChar := string(runeRequestText[len(runeRequestText)-1:])
	if strings.Index(punctuation, lastChar) < 0 {
		requestText = requestText + "？" // 判断最后字符是否加了标点，没有的话加上句号，避免openai自动补齐引起混乱。
	}

	// 4.获取上下文，拼接在一起，如果字符长度超出4000，截取为4000。（GPT按字符长度算）
	sessionText := g.service.GetUserSessionContext()
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

	// 5.返回请求文本
	return sessionText
}

// buildReply 构建回复文本
func (g *GroupMessageHandler) buildReplyText(reply string) string {
	// 1.获取@我的用户
	atText := "@" + g.sender.NickName
	textSplit := strings.Split(reply, "\n\n")
	if len(textSplit) > 1 {
		trimText := textSplit[0]
		reply = strings.Trim(reply, trimText)
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return atText + " 将物质欲望降低些，便没那么多烦恼了"
	}

	// 2.拼接回复,@我的用户，问题，回复
	replaceText := "@" + g.self.NickName
	question := strings.TrimSpace(strings.ReplaceAll(g.msg.Content, replaceText, ""))
	reply = atText + "\n" + question + "\n --------------------------------\n" + reply
	reply = strings.Trim(reply, "\n")

	// 3.返回回复的内容
	return reply
}
