package pub

import (
	"net"
	"os"
	"strconv"

	sender "github.com/chumvan/rtp/sender"
	sipClient "github.com/chumvan/sipRestServer/pkg/SIP/client"
	factoryClient "github.com/chumvan/sipRestServer/pkg/SIP/factoryClient"
)

type Publisher struct {
	FactoryClient *factoryClient.FactoryClient
	SIP           *sipClient.Client
	RTP           *sender.Sender
	RTPAddr       *net.UDPAddr
}

func New(isLocal bool) *Publisher {
	pubIP := net.ParseIP(os.Getenv("PUBLISHER_IP"))
	pubRtpPortStr := os.Getenv("PUBLISHER_RTP_PORT")
	pubRtpPort, err := strconv.Atoi(pubRtpPortStr)
	if err != nil {
		return nil
	}
	rtpAddr := &net.UDPAddr{
		IP:   pubIP,
		Port: pubRtpPort,
	}

	if isLocal {
		pub := &Publisher{
			FactoryClient: factoryClient.New(),
			SIP:           sipClient.New(),
			RTP:           &sender.Sender{},
			RTPAddr:       rtpAddr,
		}
		return pub
	} else {
		pub := &Publisher{
			FactoryClient: factoryClient.New(),
			SIP:           sipClient.New(),
			RTP:           &sender.Sender{},
			RTPAddr:       rtpAddr,
		}
		return pub
	}
}
