package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/chumvan/sipRestServer/src/server"
	"github.com/ghettovoice/gosip/log"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.InfoLevel, "Server", nil)
}

func main() {
	s := server.New()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			topic, ok := <-s.Factory.ChanTopic
			if ok {
				s.REST.CreateTopic(topic)
				break
			}
		}
	}()

	<-stop

	wg.Wait()

	s.Shutdown()
}
