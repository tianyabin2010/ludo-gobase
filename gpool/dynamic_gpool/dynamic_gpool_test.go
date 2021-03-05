package dynamic_gpool

import (
	_ "github.com/tianyabin2010/ludo-gobase/log"
	"math/rand"
	"testing"
	"time"
)

func TestGpool(*testing.T) {
	p := NewGpool("testing", 5, 5, 5)
	for i := 0; i < 10000; i++ {
		p.Post(func(){
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
		})
	}
	sig := make(chan struct{})
	<- sig
}