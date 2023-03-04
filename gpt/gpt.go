package gpt

import (
	"context"
	gogpt "github.com/sashabaranov/go-gpt3"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
)

var client *gogpt.Client

func Completions(msg []gogpt.ChatCompletionMessage) (res gogpt.ChatCompletionMessage) {
	if client == nil {
		conf := gogpt.DefaultConfig(viper.GetString("api_key"))
		if viper.GetString("proxy") != "" {
			proxyAddress, _ := url.Parse(viper.GetString("proxy"))
			conf.HTTPClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyAddress),
				},
			}
		}
		client = gogpt.NewClientWithConfig(conf)
	}
	req := gogpt.ChatCompletionRequest{
		Model:            viper.GetString("model"),
		MaxTokens:        viper.GetInt("max_tokens"),
		Temperature:      float32(viper.GetFloat64("temperature")),
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Stream:           false,
		Messages:         msg,
	}
	res = gogpt.ChatCompletionMessage{
		Role:    "system",
		Content: "将物质欲望降低些，便没那么多烦恼了",
	}
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		res.Content = err.Error()
		return
	}

	if len(resp.Choices) > 0 {
		res = resp.Choices[0].Message
	}
	return
}
