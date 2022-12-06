package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/chumvan/sipRestServer/src/pub"
	"github.com/ghettovoice/gosip/log"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.DebugLevel, "SIP-UA-Sender", nil)
}

func main() {
	wg := new(sync.WaitGroup)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Print("Start pprof on :6658\n")
		http.ListenAndServe(":6658", nil)
	}()

	logger.Info(os.Environ())

	isLocal := true
	p := pub.New(isLocal)
	if p == nil {
		logger.Error("failed to create a publisher")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := p.FactoryClient.SendRegister()
		if err != nil {
			logger.Error(err)
		}
		time.Sleep(3 * time.Second)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		topic := "amazingTopic"
		err := p.FactoryClient.InviteWithTopic(topic)
		if err != nil {
			logger.Error(err)
		}
	}()

	<-stop
	p.FactoryClient.Register.SendRegister(0)
	wg.Wait()
}
