package metrics

import (
	"testing"
	"time"
)

var (
	redisCounter = NewCounter("redis.total.count")
)

func TestMetrics(t *testing.T) {
	redisCounter.Report(1, "command=GET,status=OK")
	redisCounter.Report(1, "command=SET,status=OK")
	redisCounter.Report(1, "command=HMGET,status=OK")
	redisCounter.Report(1, "command=GET,status=OK")
	redisCounter.Report(1, "command=GET,status=OK")
	redisCounter.Report(1, "command=HMGET,status=OK")
	time.Sleep(100 * time.Second)
}
