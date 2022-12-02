package pub

import (
	SenderSIP "github.com/chumvan/sipRestServer/pkg/senderSIP"
)

type Publisher struct {
	SIP *SenderSIP.SenderSIP
}

func NewPublisher() *Publisher {
	p := &Publisher{
		SIP: SenderSIP.NewSenderSIPclient(),
	}
	return p
}
