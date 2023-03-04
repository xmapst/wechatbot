package handlers

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/viper"
	"strings"
	"time"
)

var c *cache.Cache

// MessageHandlerInterface 消息处理接口
type MessageHandlerInterface interface {
	handle() error
	ReplyText() error
}

// QrCodeCallBack 登录扫码回调，
func QrCodeCallBack(uuid string) {
	_url := fmt.Sprintf("https://login.weixin.qq.com/l/%s", uuid)
	logrus.Println("如果二维码无法扫描，请缩小控制台尺寸")
	logrus.Printf("或打开 %s\n", _url)
	q, _ := qrcode.New(_url, qrcode.High)
	fmt.Println(q.ToSmallString(true))
}

func NewHandler() (msgFunc func(msg *openwechat.Message), err error) {
	c = cache.New(viper.GetDuration("session_timeout"), time.Minute*5)
	dispatcher := openwechat.NewMessageMatchDispatcher()

	// 清空会话
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return strings.Contains(message.Content, viper.GetString("session_clear"))
	}, TokenMessageContextHandler())

	// 处理群消息
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsSendByGroup()
	}, GroupMessageContextHandler())

	// 好友申请
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsFriendAdd()
	}, func(ctx *openwechat.MessageContext) {
		msg := ctx.Message
		if viper.GetBool("auto_pass") {
			_, err := msg.Agree("")
			if err != nil {
				logrus.Warning(fmt.Sprintf("add friend agree error : %v", err))
				return
			}
		}
	})

	// 私聊
	// 获取用户消息处理器
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return !(strings.Contains(message.Content, viper.GetString("session_clear")) || message.IsSendByGroup() || message.IsFriendAdd())
	}, UserMessageContextHandler())
	return dispatcher.AsMessageHandler(), nil
}

func skipUserOrGroup(user *openwechat.User) bool {
	for _, i := range viper.GetStringSlice("ignores") {
		if strings.Contains(user.RemarkName, i) {
			return true
		}
		if strings.Contains(user.NickName, i) {
			return true
		}
		if strings.Contains(user.DisplayName, i) {
			return true
		}
	}
	return false
}
