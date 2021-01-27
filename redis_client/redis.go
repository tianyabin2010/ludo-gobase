package repo

import (
	"gitee.com/ymyy/ludo-gobase/metrics"
	"errors"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

var (
	redisTotalCommand = metrics.NewCounter("status.redis.total")
	redisLatency      = metrics.NewGauge("status.redis.latency")
)
func ReportRedis(command string, start time.Time, commandStatus string) {
	redisTotalCommand.Report(1, "method="+command+","+"commandStatus="+commandStatus)
	redisLatency.Report(float64(int64(time.Since(start))/int64(time.Millisecond)), "method="+command)
}


type redisRepo struct {
	pool *redis.Pool
}

var (
	DefaultRepo *redisRepo
	NilError    = errors.New("client is nil")
)

func InitRepo(addr string, connTimeOut, readTimeOut, writeTimeOut, maxIdle int) {
	DefaultRepo = NewRedisRepo(addr, connTimeOut, readTimeOut, writeTimeOut, maxIdle)
}

func NewRedisRepo(addr string, connTimeOut, readTimeOut, writeTimeOut, maxIdle int) *redisRepo {
	ret := &redisRepo{}
	ret.pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				addr,
				redis.DialConnectTimeout(time.Duration(connTimeOut)*time.Second),
				redis.DialReadTimeout(time.Duration(readTimeOut)*time.Second),
				redis.DialWriteTimeout(time.Duration(writeTimeOut)*time.Second),
			)
			return c, err
		},
		MaxIdle:     maxIdle,
		MaxActive:   maxIdle,
		IdleTimeout: 600 * time.Second,
		Wait:        true,
	}
	return ret
}

func (r *redisRepo) Get() redis.Conn {
	if nil != r.pool {
		return r.pool.Get()
	}
	return nil
}

func parseRedisErr(err error) string {
	if err == nil {
		return "OK"
	}
	if err.Error() == "redigo: nil returned" {
		return "OK"
	}
	switch e := err.(type) {
	case *net.OpError:
		// net error
		return e.Op + ": " + e.Err.Error()
	}
	return err.Error()
}

func (r *redisRepo) Do(command string, args ...interface{}) (interface{}, error) {
	c := r.Get()
	if c != nil {
		defer c.Close()
		start := time.Now()
		reply, err := c.Do(command, args...)
		ReportRedis(command, start, parseRedisErr(err))
		return reply, err
	}
	return nil, NilError
}

func Do(command string, args ...interface{}) (interface{}, error) {
	if nil != DefaultRepo {
		ret, err := DefaultRepo.Do(command, args...)
		if err != nil {
			log.Error().Err(err).Str("command", command).Msgf("redis exec error")
		}
		return ret, err
	}
	return nil, NilError
}
