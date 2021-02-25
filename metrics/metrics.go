package metrics

import (
	"github.com/tianyabin2010/ludo-gobase/util"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	endPoint   = ""
	reportAddr = ""
	allMetrics = map[string]MetricsReporter{} //key=name_tags...
	metricsMtx sync.RWMutex
	workCh     chan func()

	doRequest httpCallType
)

type httpCallType func(string, string, string, ...string)(*http.Response,error)

func Init(falconAddr string, httpClient httpCallType) {
	endPoint, _ = os.Hostname()
	workCh = make(chan func(), 10000)
	reportAddr = falconAddr
	doRequest = httpClient
	go run()
}

// 每 10 秒钟上报一次数据
func run() {
	defer util.BtRecover("metrics.tick")
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			reportHandler()
		case w := <-workCh:
			if nil != w {
				w()
			}
		}
	}
}

func reportHandler() {
	defer util.BtRecover("metrics.reportHandler")
	metricsMtx.RLock()
	defer metricsMtx.RUnlock()

	items := []*metricsItem{}
	for _, v := range allMetrics {
		items = append(items, v.getMetrics()...)
	}
	if len(items) <= 0 {
		return
	}
	timestamp := time.Now().Unix()
	for _, v := range items {
		v.Timestamp = timestamp
	}
	data, err := json.Marshal(items)
	if err != nil {
		log.Error().Err(err).Msgf("metrics.reportHandler marshal items error")
		return
	}
	for _, v := range allMetrics {
		v.reset()
	}
	go httpSend(data)
}

func httpSend(body []byte) {
	defer util.BtRecover("metrics.httpSend")
	resp, err := doRequest(reportAddr, "POST", string(body))
	if err != nil {
		log.Error().Err(err).
			Str("data", string(body)).
			Msgf("metrics report request error")
	}
	if nil != resp {
		defer resp.Body.Close()
	}
}

type metricsItem struct {
	Metric    string  `json:"metric"`
	Endpoint  string  `json:"endpoint"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Step      int     `json:"step"`

	// GAUGE:   即用户上传什么样的值，就原封不动的存储
	// COUNTER: 指标在存储和展现的时候，会被计算为speed，即（当前值 - 上次值）/ 时间间隔
	CounterType string `json:"counterType"`

	Tags string `json:"tags"`
}

func newMetricsItem(name, tags, tp string, val float64) *metricsItem {
	return &metricsItem{
		Metric:      name,
		Endpoint:    endPoint,
		Timestamp:   0,
		Value:       val,
		Step:        10,
		CounterType: tp,
		Tags:        tags,
	}
}

type MetricsReporter interface {
	Report(val float64, tags string)
	getMetrics() []*metricsItem
	reset()
}

type baseMetrics struct {
	name        string
	metricsType string
	metricsMap  map[string]*metricsItem
}

func (m *baseMetrics) Report(val float64, tags string) {
	workCh <- func() {
		defer util.BtRecover("metrics.baseMetrics.Report")
		if v, ok := m.metricsMap[tags]; ok {
			v.Value += val
		} else {
			item := newMetricsItem(m.name, tags, m.metricsType, val)
			m.metricsMap[tags] = item
		}
	}
}

func (m *baseMetrics) getMetrics() []*metricsItem {
	ret := make([]*metricsItem, 0, len(m.metricsMap))
	for _, v := range m.metricsMap {
		ret = append(ret, v)
	}
	return ret
}

type counterMetrics struct {
	baseMetrics
}

func (m *counterMetrics) reset() {
}

type gaugeMetrics struct {
	baseMetrics
}

func (m *gaugeMetrics) reset() {
	for _, v := range m.metricsMap {
		v.Value = 0
	}
}

func NewCounter(name string) MetricsReporter {
	ret := &counterMetrics{baseMetrics{
		name:        name,
		metricsType: "COUNTER",
		metricsMap:  make(map[string]*metricsItem),
	}}
	metricsMtx.Lock()
	defer metricsMtx.Unlock()
	allMetrics[name] = ret
	return ret
}

func NewGauge(name string) MetricsReporter {
	ret := &gaugeMetrics{baseMetrics{
		name:        name,
		metricsType: "GAUGE",
		metricsMap:  make(map[string]*metricsItem),
	}}
	metricsMtx.Lock()
	defer metricsMtx.Unlock()
	allMetrics[name] = ret
	return ret
}
