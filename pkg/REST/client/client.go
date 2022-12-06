package clientREST

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type ClientREST struct {
	ToIPAddr net.TCPAddr
	ToURL    url.URL
	client   http.Client
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

}
