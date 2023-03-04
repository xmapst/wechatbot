package main

import (
	"context"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/wechatbot/config"
	"github.com/xmapst/wechatbot/handlers"
	"path"
	"strings"
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&ConsoleFormatter{})
}

func main() {
	config.LoadConfig()
	//bot := openwechat.DefaultBot()
	bot := openwechat.NewBot(context.Background())
	openwechat.Desktop.Prepare(bot) // 桌面模式，上面登录不上的可以尝试切换这种模式

	// 扫码回调
	bot.ScanCallBack = func(_ openwechat.CheckLoginResponse) {
		logrus.Println("扫码成功,请在手机上确认登录")
	}
	// 登录回调
	bot.LoginCallBack = func(_ openwechat.CheckLoginResponse) {
		logrus.Println("登录成功")
	}

	// 注册消息处理函数
	handler, err := handlers.NewHandler()
	if err != nil {
		logrus.Fatalln("register error: %v", err)
		return
	}
	bot.MessageHandler = handler

	// 注册登陆二维码回调
	bot.UUIDCallback = handlers.QrCodeCallBack

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()
	// 执行热登录
	err = bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption())
	if err != nil {
		logrus.Errorln(err)
	}

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

type ConsoleFormatter struct {
	logrus.TextFormatter
}

func (c *ConsoleFormatter) TrimFunctionSuffix(s string) string {
	if strings.Contains(s, ".func") {
		index := strings.Index(s, ".func")
		s = s[:index]
	}
	slice := strings.Split(s, ".")
	s = slice[len(slice)-1]
	return s
}

func (c *ConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	file := path.Base(entry.Caller.File)
	function := c.TrimFunctionSuffix(path.Base(entry.Caller.Function))
	logStr := fmt.Sprintf("%s %s %s:%d %s %v\n",
		entry.Time.Format("2006/01/02 15:04:05"),
		strings.ToUpper(entry.Level.String()),
		file,
		entry.Caller.Line,
		function,
		entry.Message,
	)
	return []byte(logStr), nil
}
