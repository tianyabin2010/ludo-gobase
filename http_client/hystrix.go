package httpclient

import (
	"github.com/afex/hystrix-go/hystrix"
	"time"
)

var DefaultCircutBreakerConfig = hystrix.CommandConfig{
	Timeout:                4000,
	MaxConcurrentRequests:  2048,
	RequestVolumeThreshold: 20,
	SleepWindow:            5000,
	ErrorPercentThreshold:  30,
}

func init() {
	hystrix.DefaultTimeout = 4000
	hystrix.DefaultMaxConcurrent = 2048
	hystrix.DefaultVolumeThreshold = 100
	hystrix.DefaultSleepWindow = 5000
	hystrix.DefaultErrorPercentThreshold = 50
}

//ConfigureCircuitBreaker ...
func ConfigureCircuitBreaker(key string, cfg hystrix.CommandConfig) {
	hystrix.ConfigureCommand(key, cfg)
}

func getTimeout(key string) time.Duration {
	settings := hystrix.GetCircuitSettings()
	if setting, ok := settings[key]; ok {
		return setting.Timeout
	}
	return time.Duration(DefaultCircutBreakerConfig.Timeout) * time.Millisecond
}
