package httpclient

import (
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	cliPool = newClientPool()
)

type clientPool struct {
	sync.RWMutex
	clients map[string]*http.Client
}

func newClientPool() *clientPool {
	return &clientPool{
		clients: make(map[string]*http.Client),
	}
}

func (p *clientPool) getClient(host string) *http.Client {
	p.RLock()
	if client, ok := p.clients[host]; ok {
		p.RUnlock()
		return client
	}
	p.RUnlock()
	client := &http.Client{
		Transport: &ochttp.Transport{
			Base: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConnsPerHost:   128,
				MaxIdleConns:          100,
				IdleConnTimeout:       time.Second * 90,
				ExpectContinueTimeout: 5 * time.Second, //目前应该没有请求使用了Expect: 100-continue
			},
			Propagation: &b3.HTTPFormat{},
		},
	}
	p.Lock()
	p.clients[host] = client
	p.Unlock()
	return client
}
