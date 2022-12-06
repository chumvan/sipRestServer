package clientREST

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/ghettovoice/gosip/log"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.InfoLevel, "Conference Factory", nil)
}

type ClientREST struct {
	ToIPAddr  net.TCPAddr
	ToURL     url.URL
	client    http.Client
	ChanTopic chan string
}

func NewClientREST(to *net.TCPAddr) (cr *ClientREST, err error) {
	cr = &ClientREST{
		ToIPAddr: *to,
	}
	toUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", to.IP, to.Port))
	if err != nil {
		return nil, err
	}
	cr.ToURL = *toUrl
	cr.client = http.Client{Timeout: time.Duration(1) * time.Second}
	return cr, nil
}

func (cr *ClientREST) CreateTopic(topic string) {
	logger.Infof("received topic: %s", topic)
}
