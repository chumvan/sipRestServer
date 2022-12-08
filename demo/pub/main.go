package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/chumvan/rtp/scraper"
	"github.com/chumvan/sipRestServer/src/pub"
	"github.com/chumvan/t140/t140packet"
	"github.com/ghettovoice/gosip/log"
	"github.com/pion/rtp"
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

	// REST server
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Print("Start pprof on :6658\n")
		http.ListenAndServe(":6658", nil)
	}()

	logger.Info(os.Environ())

	isLocal := true
	publisher := pub.New(isLocal)
	if publisher == nil {
		logger.Error("failed to create a publisher")
	}

	// Send REGISTER
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := publisher.FactoryClient.SendRegister()
		if err != nil {
			logger.Error(err)
		}
		time.Sleep(3 * time.Second)
	}()

	// Send INVITE with topic in sdp body
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Second)
		topic := "amazingTopic"
		err := publisher.FactoryClient.InviteWithTopic(topic)
		if err != nil {
			logger.Error(err)
		}
	}()

	// start RTP streaming when target (forwarder) IP passed to channel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case addr := <-publisher.FactoryClient.TargetChan:

				targetAddr, err := net.ResolveUDPAddr("udp4", addr)
				if err != nil {
					logger.Error(err)
				}
				logger.Infof("start streaming to address: %s\n", targetAddr)

				ticker := time.NewTicker(1000 * time.Millisecond)
				done := make(chan bool)
				// RTP stream start
				packetizer := rtp.NewPacketizer(
					1500,
					100,
					5000,
					&t140packet.T140Payloader{},
					rtp.NewRandomSequencer(),
					1000,
				)
				scraper := &scraper.Scraper{}
				err = publisher.RTP.DialTo(publisher.RTPAddr, targetAddr)
				if err != nil {
					logger.Error(err)
				}

				go func() {
					for {
						select {
						case <-done:
							publisher.RTP.Close()
							return
						case t := <-ticker.C:
							payload, err := scraper.Scrape()
							if err != nil {
								panic(err)
							}
							packets := packetizer.Packetize(payload, 1)

							for _, p := range packets {
								rawBuf, err := p.Marshal()
								if err != nil {
									panic(err)
								}
								publisher.RTP.Send(rawBuf)
								fmt.Println("packet sent at: ", t)
							}
						}
					}
				}()

				// RTP stream end
				time.Sleep(300000 * time.Millisecond)
				ticker.Stop()
				done <- true
				logger.Info("finished all sending")

			case <-time.After(3 * time.Second):
				logger.Error("no address received")

			}
		}
	}()

	<-stop
	publisher.FactoryClient.Register.SendRegister(0)
	wg.Wait()
}
