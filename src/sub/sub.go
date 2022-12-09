package sub

import (
	"net"
	"os"
	"strconv"

	"github.com/chumvan/rtp/receiver"
)

type Subscriber struct {
	RTP     *receiver.Receiver
	RTPAddr *net.UDPAddr
}

func New() *Subscriber {
	subIP := net.ParseIP(os.Getenv("SUB_IP"))
	subRtpPortStr := os.Getenv("SUB_RTP_PORT")
	subRtpPort, err := strconv.Atoi(subRtpPortStr)
	if err != nil {
		return nil
	}
	rtpAddr := &net.UDPAddr{
		IP:   subIP,
		Port: subRtpPort,
	}
	sub := &Subscriber{
		RTP:     &receiver.Receiver{},
		RTPAddr: rtpAddr,
	}
	return sub
}
