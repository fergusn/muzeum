package events

import (
	"sync"
	"testing"
)

func TestAllSubscribersReceiveEvent(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	for i := 0; i < 2; i++ {
		go func() {
			<-Package.Pulled.Receive()
			wg.Done()
		}()
	}

	for len(Package.Pulled.subscribers) < 2 {
	}

	Package.Pulled.Emit(&Pulled{})

	wg.Wait()

}
