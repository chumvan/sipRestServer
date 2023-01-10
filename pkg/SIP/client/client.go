package client

import (
	"net"
	"os"
	"strconv"

	"github.com/chumvan/go-sip-ua/pkg/ua"
	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/chumvan/sipRestServer/src/mock"
	"github.com/cloudwebrtc/go-sip-ua/pkg/account"
	"github.com/cloudwebrtc/go-sip-ua/pkg/session"
	"github.com/cloudwebrtc/go-sip-ua/pkg/stack"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
	"github.com/pion/sdp"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.DebugLevel, "SIP-UAC", nil)
}

type Client struct {
	UDPAddress *net.UDPAddr

	UA    *ua.UserAgent
	Stack *stack.SipStack

	TargetIP   string
	TargetChan chan string
	TargetPort int
	Recipient  sip.SipUri

	serverIP   string
	serverPort int

	Profile  *account.Profile
	Register *ua.Register
}

func New() (s *Client) {
	s = &Client{}
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

	s.TargetChan = make(chan string, 1)

	s.UA.InviteStateHandler = func(sess *session.Session, req *sip.Request, resp *sip.Response, state session.Status) {
		logger.Infof("InviteStateHandler: state => %v, type => %s", state, sess.Direction())

		switch state {
		case session.InviteReceived:

			sdp := mock.BuildLocalSdp(senderIPstr, senderPort)
			logger.Infof("received INVITE")
			sess.ProvideAnswer(sdp)
			sess.Accept(200)
		case session.Confirmed:
			logger.Infof("confirmed INVITE")
			localSdp := sess.RemoteSdp()
			sessionDesc := &sdp.SessionDescription{}
			err := sessionDesc.Unmarshal(localSdp)
			if err != nil {
				logger.Error(err)
			}

			confUri, _ := sessionDesc.Attribute("confUri")
			topicIP, _ := sessionDesc.Attribute("topicIP")
			topicPort, _ := sessionDesc.Attribute("topicPort")
			logger.Infof(`confirmed conference with:
					 		conference uri: %s
							at topic IP: %s
							at topic Port: %s`,
				confUri,
				topicIP,
				topicPort)

			s.TargetIP = topicIP
			s.TargetPort, err = strconv.Atoi(topicPort)
			if err != nil {
				logger.Error(err)
			}
			// pass the target's (forwarder's) IP for data transferring
			out := topicIP + ":" + topicPort
			s.TargetChan <- out

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
	var serverIP string
	serverIP, ok = os.LookupEnv("SERVER_SIP_IP")
	if !ok {
		logger.Error("serverSipIP param not found")
	}

	var serverPort string

	serverPort, ok = os.LookupEnv("SERVER_SIP_PORT")
	if !ok {
		logger.Error("serverSipPort param not found")
	}

	sipUriString := "sip" + ":" +
		"server" + "@" + serverIP + ":" +
		serverPort + ";" +
		"transport=udp"

	s.Recipient, err = parser.ParseSipUri(sipUriString)
	if err != nil {
		logger.Error(err)
	}

	s.serverIP = serverIP
	s.serverPort, err = strconv.Atoi(serverPort)
	if err != nil {
		logger.Error(err)
	}

	return s
}

func (s *Client) SendRegister() (err error) {
	s.Register, err = s.UA.SendRegister(s.Profile, s.Recipient, s.Profile.Expires, nil)
	return
}
