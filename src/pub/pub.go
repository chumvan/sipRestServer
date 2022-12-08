package pub

import (
	"net"
	"os"
	"strconv"

	sender "github.com/chumvan/rtp/sender"
	senderSIP "github.com/chumvan/sipRestServer/pkg/SIP/sender"
)

type Publisher struct {
	FactoryClient *senderSIP.SenderSIP
	SIP           *senderSIP.SenderSIP
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
		p := &Publisher{
			FactoryClient: senderSIP.NewSenderSIPclient("local-factory"),
			SIP:           senderSIP.NewSenderSIPclient("local-server"),
			RTP:           &sender.Sender{},
			RTPAddr:       rtpAddr,
		}
		return p
	} else {
		p := &Publisher{
			FactoryClient: senderSIP.NewSenderSIPclient("factory"),
			SIP:           senderSIP.NewSenderSIPclient("server"),
			RTP:           &sender.Sender{},
			RTPAddr:       rtpAddr,
		}
		return p
	}
}
