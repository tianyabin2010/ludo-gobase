package httpclient

import (
	"github.com/rs/zerolog/log"
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
				Dial: func(network, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(network, addr, time.Second*3)
					if err != nil {
						return nil, err
					}
					log.Info().Str("addr", addr).Msg("Create one connection")
					return c, nil
				},
				MaxIdleConnsPerHost:   128,
				MaxIdleConns:          2048,
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
