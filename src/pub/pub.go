package pub

import (
	SenderSIP "github.com/chumvan/sipRestServer/pkg/SIP/sender"
)

type Publisher struct {
	FactoryClient *SenderSIP.SenderSIP
	SIP           *SenderSIP.SenderSIP
}

func New(isLocal bool) *Publisher {
	if isLocal {
		p := &Publisher{
			FactoryClient: SenderSIP.NewSenderSIPclient("local-factory"),
			SIP:           SenderSIP.NewSenderSIPclient("local-server"),
		}
		return p
	} else {
		p := &Publisher{
			FactoryClient: SenderSIP.NewSenderSIPclient("factory"),
			SIP:           SenderSIP.NewSenderSIPclient("server"),
		}
		return p
	}
}
