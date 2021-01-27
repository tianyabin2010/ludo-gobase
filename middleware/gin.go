package middleware

import (
	"gitee.com/ymyy/ludo-gobase/metrics"
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	ginlimiter "github.com/julianshen/gin-limiter"
	"github.com/rs/zerolog/log"
)

var (
	statusHttpMetrics = metrics.NewCounter("status.http.totalcount")
	statusHttpLatency = metrics.NewGauge("status.http.latency")
)

func reportMetrics(method string, status int, start time.Time) {
	statusHttpMetrics.Report(1, fmt.Sprintf("method=%v,status=%v", method, status))
	statusHttpLatency.Report(float64(int64(time.Since(start))/int64(time.Millisecond)), fmt.Sprintf("method=%v", method))
}

type respLogWriter struct {
	gin.ResponseWriter
	resp *bytes.Buffer
}

func (r respLogWriter) Write(b []byte) (int, error) {
	r.resp.Write(b)
	return r.ResponseWriter.Write(b)
}

// NewRateLimiter ...
func NewRateLimiter(interval time.Duration, cap int64) gin.HandlerFunc {
	return ginlimiter.NewRateLimiter(interval, cap, func(ctx *gin.Context) (string, error) {
		return "", nil
	}).Middleware()
}

// NewLogWrapper ...
func NewLogWrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		rw := &respLogWriter{
			resp:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = rw
		b, _ := c.GetRawData()
		c.Request.Body.Close()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		req := map[string]interface{}{}
		req["method"] = c.Request.Method
		req["header"] = c.Request.Header
		req["param"] = c.Params
		req["body"] = string(b)
		begin := time.Now()
		c.Next()
		end := time.Now()
		log.Debug().
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Interface("request", req).
			Str("response", rw.resp.String()).
			TimeDiff("cost", end, begin).
			Msg("gin.HandleFunc")
		reportMetrics(c.Request.Method, c.Writer.Status(), begin)
	}
}
