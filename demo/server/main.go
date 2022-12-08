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
			topicMeta, ok := <-s.Factory.ChanMeta
			if ok {
				logger.Infof("topic meta: %s", topicMeta)
				s.REST.CreateTopic(topicMeta, s.Factory.ChanInfo)
				break
			} else {
				logger.Error("no topic meta returned")
				break
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			topicInfo, ok := <-s.Factory.ChanInfo
			if ok {
				logger.Infof("topic info: %s", topicInfo)
				// pass to topicInfo channel of the factory
				s.Factory.ChanInfo <- topicInfo
				break
			} else {
				logger.Error("No topic info return")
			}
		}
	}()

	<-stop

	wg.Wait()

	s.Shutdown()
}
