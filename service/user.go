package service

import (
	"github.com/eatmoreapple/openwechat"
	"github.com/patrickmn/go-cache"
	gogpt "github.com/sashabaranov/go-gpt3"
	"github.com/spf13/viper"
	"time"
)

// UserServiceInterface 用户业务接口
type UserServiceInterface interface {
	GetUserSessionContext() []gogpt.ChatCompletionMessage
	SetUserSessionContext(reply []gogpt.ChatCompletionMessage)
	ClearUserSessionContext()
}

var _ UserServiceInterface = (*UserService)(nil)

// UserService 用戶业务
type UserService struct {
	// 缓存
	cache *cache.Cache
	// 用户
	user *openwechat.User
}

// NewUserService 创建新的业务层
func NewUserService(cache *cache.Cache, user *openwechat.User) UserServiceInterface {
	return &UserService{
		cache: cache,
		user:  user,
	}
}

// ClearUserSessionContext 清空GTP上下文，接收文本中包含`我要问下一个问题`，并且Unicode 字符数量不超过20就清空
func (s *UserService) ClearUserSessionContext() {
	s.cache.Delete(s.user.ID())
}

// GetUserSessionContext 获取用户会话上下文文本
func (s *UserService) GetUserSessionContext() []gogpt.ChatCompletionMessage {
	// 1.获取上次会话信息，如果没有直接返回空字符串
	sessionContext, ok := s.cache.Get(s.user.ID())
	if !ok {
		return nil
	}

	// 2.返回上文
	return sessionContext.([]gogpt.ChatCompletionMessage)
}

// SetUserSessionContext 设置用户会话上下文文本，question用户提问内容，GTP回复内容
func (s *UserService) SetUserSessionContext(reply []gogpt.ChatCompletionMessage) {
	s.cache.Set(s.user.ID(), reply, time.Second*viper.GetDuration("session_timeout"))
}
