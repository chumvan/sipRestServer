package serverSIP

import (
	"fmt"
	"net"
	"os"

	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/cloudwebrtc/go-sip-ua/pkg/account"
	"github.com/cloudwebrtc/go-sip-ua/pkg/session"
	"github.com/cloudwebrtc/go-sip-ua/pkg/stack"
	"github.com/cloudwebrtc/go-sip-ua/pkg/ua"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/transport"
)

type SIPServer struct {
	UDPAddress *net.UDPAddr
	Stack      *stack.SipStack
	UA         *ua.UserAgent
}

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.InfoLevel, "SIP Server", nil)
}

func New() *SIPServer {
	ss := &SIPServer{}
	stack := stack.NewSipStack(&stack.SipStackConfig{
		UserAgent:  "SIP Server",
		Extensions: []string{"replace", "outbound"},
		Dns:        "8.8.8.8",
	})

	stack.OnConnectionError(ss.handleConnectionError)
	serverIP, ok := os.LookupEnv("SERVER_SIP_IP")
	if !ok {
		logger.Error("ServerSipIP param not found")
	}
	portStr, ok := os.LookupEnv("SERVER_SIP_PORT")
	if !ok {
		logger.Error("confFactoryPort param not found")
	}
	listen := fmt.Sprint(serverIP) + ":" + portStr
	var err error
	ss.UDPAddress, err = net.ResolveUDPAddr("udp", listen)
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
			to, _ := (*req).To()
			from, _ := (*req).From()
			caller := from.Address
			called := to.Address

			offer := sess.RemoteSdp()
			logger.Infof("from: %v to: %v \nsdp: %v", caller, called, offer)
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

	stack.OnRequest(sip.REGISTER, ss.handleRegister)
	stack.OnRequest(sip.BYE, ss.handleBye)
	ss.Stack = stack
	ss.UA = ua

	return ss
}

func (ss *SIPServer) handleBye(req sip.Request, tx sip.ServerTransaction) {
	logger.Info("BYE received, %v\n recipient", req.String(), req.Recipient())
}

func (ss *SIPServer) handleRegister(req sip.Request, tx sip.ServerTransaction) {
	headers := req.GetHeaders("Expires")
	to, _ := req.To()
	var expires sip.Expires = 0
	if len(headers) > 0 {
		expires = *headers[0].(*sip.Expires)
	}
	reason := ""
	if len(headers) > 0 && expires != sip.Expires(0) {
		logger.Infof("Registered [%v] expires [%d] source %s", to, expires, req.Source())
		reason = "Registered"
	} else {
		logger.Infof("Logged out [%v] expires [%d] ", to, expires)
		reason = "UnRegistered"
	}
	resp := sip.NewResponseFromRequest(req.MessageID(), req, 200, reason, "")
	sip.CopyHeaders("Expires", req, resp)
	utils.BuildContactHeader("Contact", req, resp, &expires)
	tx.Respond(resp)

}
func (ss *SIPServer) handleConnectionError(connError *transport.ConnectionError) {
	logger.Debugf("Handle Connection Lost: Source: %v, Dest: %v, Network: %v", connError.Source, connError.Dest, connError.Net)
}

func (ss *SIPServer) SetLogLevel(level log.Level) {
	utils.SetLogLevel("SIP Server", level)
}

func (ss *SIPServer) Shutdown() {
	ss.UA.Shutdown()
}
