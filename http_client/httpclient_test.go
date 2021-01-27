package httpclient

import (
	"testing"
	"time"
)

func TestHystrix(t *testing.T) {
	tmp := []int{1, 2, 3, 4}
	blocks := [][]int{}
	blockNum := len(tmp) / 5
	surplus := len(tmp) % 5
	for i := 0; i < blockNum; i++ {
		blocks = append(blocks, tmp[i*5:i*5+5])
	}
	if surplus > 0 {
		blocks = append(blocks, tmp[blockNum*5:])
	}
	DoRequest("http://127.0.0.1:9111/status/v1/query", "POST", "{\"userIds\":[123]}")
	time.Sleep(100 * time.Second)
}
