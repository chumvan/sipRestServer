package server

import (
	"net"
	"os"
	"strconv"

	clientREST "github.com/chumvan/sipRestServer/pkg/REST/client"
	conffactory "github.com/chumvan/sipRestServer/pkg/SIP/factory"
	serverSIP "github.com/chumvan/sipRestServer/pkg/SIP/server"
)

type Server struct {
	Factory *conffactory.ConfFactory
	SIP     *serverSIP.SIPServer
	REST    *clientREST.ClientREST
}

func New() *Server {
	f := conffactory.New()
	if f == nil {
		return nil
	}
	restServerIP := os.Getenv("CONF_TOPIC_MAPPER_IP")
	restServerPort, _ := strconv.Atoi(os.Getenv("CONF_TOPIC_MAPPER_PORT"))
	restServerAddr := &net.TCPAddr{
		IP:   net.ParseIP(restServerIP),
		Port: restServerPort,
	}
	r := clientREST.New(restServerAddr, f.UDPAddress)
	if r == nil {
		return nil
	}
	s := &Server{
		Factory: f,
		REST:    r,
	}
	return s
}

func (s *Server) Shutdown() {
	s.Factory.Shutdown()
	s.SIP.Shutdown()
}
