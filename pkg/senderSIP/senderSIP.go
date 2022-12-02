package SenderSIP

import (
	"net"
	"os"
	"strconv"

	"github.com/chumvan/go-sip-ua/examples/mock"
	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/cloudwebrtc/go-sip-ua/pkg/account"
	"github.com/cloudwebrtc/go-sip-ua/pkg/session"
	"github.com/cloudwebrtc/go-sip-ua/pkg/stack"
	"github.com/cloudwebrtc/go-sip-ua/pkg/ua"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
	"github.com/pion/sdp"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.DebugLevel, "SIP-UA-Sender", nil)
}

type SenderSIP struct {
	UDPAddress *net.UDPAddr

	UA    *ua.UserAgent
	Stack *stack.SipStack

	targetIP   string
	targetChan chan string
	Recipient  sip.SipUri

	Profile  *account.Profile
	Register *ua.Register
}

func NewSenderSIPclient() (s *SenderSIP) {
	s = &SenderSIP{}
	senderIPstr, ok := os.LookupEnv("SENDER_IP")
	if !ok {
		logger.Error("senderIP param not found")
	}
	senderPortStr, ok := os.LookupEnv("SENDER_PORT")
	if !ok {
		logger.Error("senderPort param not found")
	}
	senderPort, err := strconv.Atoi(senderPortStr)
	if err != nil {
		logger.Error(err)
	}

	senderName, ok := os.LookupEnv("SENDER_NAME")
	if !ok {
		logger.Error("localName param not found")
	}

	service := senderIPstr + ":" + senderPortStr
	s.UDPAddress, err = net.ResolveUDPAddr("udp", service)
	if err != nil {
		logger.Error(err)
	}

	localIPAddr := senderIPstr + ":" + senderPortStr
	logger.Infof("listen to %s", localIPAddr)

	s.Stack = stack.NewSipStack(&stack.SipStackConfig{
		UserAgent:  "Sender SIP UAC",
		Extensions: []string{"replaces", "outbound"},
		Dns:        "8.8.8.8"})
	if err := s.Stack.Listen("udp", localIPAddr); err != nil {
		logger.Error(err)
	}

	s.UA = ua.NewUserAgent(&ua.UserAgentConfig{
		SipStack: s.Stack,
	})

	s.targetChan = make(chan string, 1)

	s.UA.InviteStateHandler = func(sess *session.Session, req *sip.Request, resp *sip.Response, state session.Status) {
		logger.Infof("InviteStateHandler: state => %v, type => %s", state, sess.Direction())

		switch state {
		case session.InviteReceived:

			sdp := mock.BuildLocalSdp(senderIPstr, senderPort)
			logger.Infof("Received INVITE")
			sess.ProvideAnswer(sdp)
			sess.Accept(200)
		case session.Confirmed:
			logger.Infof("Confirmed INVITE")
			localSdp := sess.RemoteSdp()
			sessionDesc := &sdp.SessionDescription{}
			err := sessionDesc.Unmarshal(localSdp)
			if err != nil {
				logger.Error(err)
			}
			s.targetIP = sessionDesc.Origin.UnicastAddress
			logger.Infof("receiverIP = %s", s.targetIP)
			// Signaling is done, go for data transferring
			s.targetChan <- s.targetIP

		case session.Canceled:
			fallthrough
		case session.Failure:
			fallthrough
		case session.Terminated:
			logger.Info("Session terminated")
		}
	}

	s.UA.RegisterStateHandler = func(state account.RegisterState) {
		logger.Infof("RegisterStateHandler: user => %s, state => %v, expires => %v", state.Account.AuthInfo.AuthUser, state.StatusCode, state.Expiration)
	}

	// UAC uri
	uriString := "sip:" + senderName + "@" + s.UDPAddress.String()
	senderUri, err := parser.ParseUri(uriString)
	if err != nil {
		logger.Error(err)
	}

	// A profile for each Participant
	s.Profile = account.NewProfile(senderUri.Clone(), "RPi Sender",
		&account.AuthInfo{
			AuthUser: utils.SenderName,
			Password: utils.SenderPass,
			Realm:    "",
		},
		utils.DefaultExpires,
		s.Stack)

	// A recipient = SIP server
	serverIP, ok := os.LookupEnv("SERVER_IP")
	if !ok {
		logger.Error("serverIP param not found")
	}
	serverPort, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		logger.Error("serverPort param not found")
	}
	sipUriString := "sip" + ":" +
		senderName + "@" + serverIP + ":" +
		serverPort + ";" +
		"transport=udp"
	s.Recipient, err = parser.ParseSipUri(sipUriString)
	if err != nil {
		logger.Error(err)
	}

	return s
}

func (s *SenderSIP) SendRegister() (err error) {
	s.Register, err = s.UA.SendRegister(s.Profile, s.Recipient, s.Profile.Expires, nil)
	return
}
