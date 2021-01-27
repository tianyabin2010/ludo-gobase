package httpclient

import (
	"gitee.com/ymyy/ludo-gobase/metrics"
	"bytes"
	"context"
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var (
	E_GetClient = errors.New("get client error")
)

var (
	httpRequestMetrics = metrics.NewCounter("stat.http.total")
	httpRequestLatency = metrics.NewGauge("stat.http.latency")
)

// Param.headers 必须是偶数, 奇数为key, 偶数val
func DoRequest(url, method, body string, headers ...string) (*http.Response, error) {
	var err error
	defer func(start time.Time) {
		status := "OK"
		if err != nil {
			status = err.Error()
		}
		httpRequestMetrics.Report(1, "url="+url+",method="+method+",status="+status)
		httpRequestLatency.Report(float64(int64(time.Since(start))/int64(time.Millisecond)), "url="+url+",method="+method)
	}(time.Now())
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(body)))
	if err != nil {
		log.Error().Err(err).
			Str("url", url).
			Str("method", method).
			Str("body", body).
			Msgf("httpclientDoRequest new request error")
		return nil, err
	}
	if len(headers) > 0 && len(headers)%2 == 0 {
		for i := 0; i < len(headers); i += 2 {
			req.Header.Set(headers[i], headers[i+1])
		}
	}
	cli := cliPool.getClient(url)
	if nil == cli {
		err = E_GetClient
		return nil, err
	}

	var resp *http.Response
	ctx, cancel := context.WithCancel(context.Background())
	hystridCh := hystrix.GoC(context.Background(), url, func(cancelFn func()) func(context.Context) error {
		return func(ctx context.Context) error {
			resp, err = cli.Do(req)
			if err != nil {
				return err
			}
			cancelFn()
			return nil
		}
	}(cancel), nil)
	select {
	case err = <-hystridCh:
		if err != nil {
			log.Error().Err(err).
				Str("url", url).
				Str("method", method).
				Str("body", body).
				Msgf("httpclientDoRequest request error")
			return nil, err
		}
	case <-ctx.Done():
		if resp == nil {
			return nil, errors.New("resp is nil")
		}
		return resp, nil
	}
	return resp, nil
}
