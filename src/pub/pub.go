package pub

import (
	SenderSIP "github.com/chumvan/sipRestServer/pkg/senderSIP"
)

type Publisher struct {
	SIP SenderSIP.SenderSIP
}

func NewPublisher(p *Publisher) {
	p.SIP = *SenderSIP.NewSenderSIPclient()
}
