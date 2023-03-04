package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	pflag.String("api_key", "", "Chat GPT API Token")
	pflag.Bool("auto_pass", true, "自动通过好友")
	pflag.Duration("session_timeout", 15*time.Second, "会话超时时间")
	pflag.Int("max_tokens", 512, "GPT请求最大字符数")
	pflag.String("model", "text-davinci-003", "GPT模型")
	pflag.Float64("temperature", 1.0, "热度")
	pflag.String("reply_prefix", "", "回复前缀")
	pflag.String("session_clear", "下一个问题", "清空会话口令")
	pflag.StringSlice("ignores", nil, "忽略包含关键字的群组或用户")
	pflag.String("proxy", "", "GPT使用http(s)/socks(5)代理")
	pflag.String("conf", "config.json", "配置文件路径")
	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)
}

// LoadConfig 加载配置
func LoadConfig() {
	viper.SetConfigName(filepath.Base(viper.GetString("conf")))
	viper.AddConfigPath(filepath.Dir(viper.GetString("conf")))
	viper.SetConfigType("json")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalln(err)
	}
	if viper.GetInt64("max_tokens") >= 2048 {
		viper.Set("max_tokens", 2048)
	}
}
