package examples

import (
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestXXTRead(t *testing.T) {
	for {
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
		rateStop := rand.Intn(100)
		if rateStop < 30 {
			time.Sleep(time.Duration(rand.Intn(10)+10) * time.Second)
		}
		//submitReadLog()
		log.Println("action1")

	}
}
