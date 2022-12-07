package conffactory

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/cloudwebrtc/go-sip-ua/pkg/account"
	"github.com/cloudwebrtc/go-sip-ua/pkg/session"
	"github.com/cloudwebrtc/go-sip-ua/pkg/stack"
	"github.com/cloudwebrtc/go-sip-ua/pkg/ua"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/transport"
	"github.com/pixelbender/go-sdp/sdp"
)

type ConfFactory struct {
	UDPAddress *net.UDPAddr
	Stack      *stack.SipStack
	UA         *ua.UserAgent
	ChanTopic  chan string
}

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.InfoLevel, "Conference Factory", nil)
}

func New() *ConfFactory {
	cf := &ConfFactory{}
	cf.ChanTopic = make(chan string, 1)

	stack := stack.NewSipStack(&stack.SipStackConfig{
		UserAgent:  "Conference Factory",
		Extensions: []string{"replace", "outbound"},
		Dns:        "8.8.8.8",
	})

	stack.OnConnectionError(cf.handleConnectionError)

	localIP, err := utils.GetOutboundIP()
	if err != nil {
		logger.Error(err)
	}
	factoryIP, ok := os.LookupEnv("SERVER_SIP_IP")
	if !ok {
		logger.Error("serverSipIP param not found")
	}

	confFactoryPortStr, ok := os.LookupEnv("CONF_FACTORY_PORT")
	if !ok {
		logger.Error("confFactoryPort param not found")
	}
	factoryPort, err := strconv.Atoi(confFactoryPortStr)
	if err != nil {
		logger.Error(err)
	}

	listen := fmt.Sprint(localIP) + ":" + confFactoryPortStr
	cf.UDPAddress, err = net.ResolveUDPAddr("udp", listen)
	if err != nil {
		logger.Error(err)
	}

	if err := stack.Listen("udp", listen); err != nil {
		logger.Panic(err)
	}

	ua := ua.NewUserAgent(&ua.UserAgentConfig{
		SipStack: stack,
	})

	ua.InviteStateHandler = func(sess *session.Session, req *sip.Request, resp *sip.Response, state session.Status) {
		logger.Infof("InviteStateHandler: state => %v, type => %s", state, sess.Direction())

		switch state {
		// Handle incoming call.
		case session.InviteReceived:
			// to, _ := (*req).To()
			// from, _ := (*req).From()
			// // caller := from.Address
			// called := to.Address

			offer := sess.RemoteSdp()

			// logger.Infof("from: %v to: %v \nsdp: %v", caller, called, offer)
			// append an attribute to send back a conference URI
			tempSessDesc, err := sdp.ParseString(offer)
			if err != nil {
				logger.Error(err)
			}

			topicName := tempSessDesc.Attributes.Get("topic")
			// logger.Infof("topic name: %s", topicName)
			cf.ChanTopic <- topicName

			confUri := fmt.Sprintf("sip:%s@%s:%d;transport=udp", topicName, factoryIP, factoryPort)
			forwarderIP := os.Getenv("FORWARDER_IP")
			forwarderRtpInPort := os.Getenv("FORWARDER_RTP_IN_PORT")
			tempSessDesc.Attributes = append(tempSessDesc.Attributes,
				sdp.NewAttr("confUri", confUri),
				sdp.NewAttr("topicIP", forwarderIP),
				sdp.NewAttr("topicPort", forwarderRtpInPort))
			answer := tempSessDesc.String()
			sess.ProvideAnswer(answer)
			sess.Accept(200)
			return
		// Handle re-INVITE or UPDATE.
		case session.ReInviteReceived:
			logger.Infof("re-INVITE")
			switch sess.Direction() {
			case session.Incoming:
				sess.Accept(200)
			case session.Outgoing:
			}
			return
		// Handle 1XX
		case session.EarlyMedia:
			logger.Infof("EARLYMEDIA")
			return
		case session.Provisional:
			logger.Infof("PROVISIONAL")
			return
		// Handle 200OK or ACK
		case session.Confirmed:
			logger.Infof("CONFIRMED")
			return
			// Handle 4XX+
		case session.Failure:
			logger.Infof("FAILURE")
			return
		case session.Canceled:
			logger.Infof("CANCELED")
			return
		case session.Terminated:
			logger.Info("Terminated")
			return
		}
	}

	ua.RegisterStateHandler = func(state account.RegisterState) {
		logger.Infof("RegisterStateHandler: state => %v", state)
	}

	stack.OnRequest(sip.REGISTER, cf.handleRegister)
	cf.Stack = stack
	cf.UA = ua

	return cf
}

func (cf *ConfFactory) handleRegister(request sip.Request, tx sip.ServerTransaction) {
	headers := request.GetHeaders("Expires")
	to, _ := request.To()
	var expires sip.Expires = 0
	if len(headers) > 0 {
		expires = *headers[0].(*sip.Expires)
	}
	reason := ""
	if len(headers) > 0 && expires != sip.Expires(0) {
		logger.Infof("Registered [%v] expires [%d] source %s", to, expires, request.Source())
		reason = "Registered"
	} else {
		logger.Infof("Logged out [%v] expires [%d] ", to, expires)
		reason = "UnRegistered"
	}
	resp := sip.NewResponseFromRequest(request.MessageID(), request, 200, reason, "")
	sip.CopyHeaders("Expires", request, resp)
	utils.BuildContactHeader("Contact", request, resp, &expires)
	tx.Respond(resp)

}

func (cf *ConfFactory) handleConnectionError(connError *transport.ConnectionError) {
	logger.Debugf("Handle Connection Lost: Source: %v, Dest: %v, Network: %v", connError.Source, connError.Dest, connError.Net)
}

func (cf *ConfFactory) SetLogLevel(level log.Level) {
	utils.SetLogLevel("ConfFactory", level)
}

func (cf *ConfFactory) Shutdown() {
	cf.UA.Shutdown()
}
