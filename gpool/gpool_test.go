package gpool

import (
	_ "gitee.com/ymyy/ludo-gobase/log"
	"fmt"
	"testing"
	"time"
)

func TestGpool(t *testing.T) {
	pool := NewGpool("test_gpool", 10)
	i := 0
	for {
		pool.Post(func(num int) Job {
			timeStr := time.Now().Format("2006-01-02 15:04:05.000")
			return func() {
				if num < 10 {
					time.Sleep(time.Millisecond * 20)
				}
				fmt.Println(timeStr, "[", num, "]")
			}
		}(i))
		i++
		if i >= 1000 {
			time.Sleep(100 * time.Second)
		}
	}
}