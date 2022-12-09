package main

import (
	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/chumvan/sipRestServer/src/sub"
	"github.com/ghettovoice/gosip/log"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.DebugLevel, "UE-SUBSCRIBER", nil)
}

func main() {
	sub := sub.New()
	if sub == nil {
		logger.Error("failed to create subscriber")
		return
	}

	err := sub.RTP.ListenOn(sub.RTPAddr)
	if err != nil {
		logger.Error(err)
	}
	defer sub.RTP.Close()
	sub.RTP.Loop()
	logger.Info("finished program")
}
