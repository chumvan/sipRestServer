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
	SIP     *serverSIP.ServerSIP
	REST    *clientREST.ClientREST
}

func New() *Server {
	f := conffactory.New()
	if f == nil {
		return nil
	}
	toIP := os.Getenv("DB_IP")
	toPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	to := &net.TCPAddr{
		IP:   net.IP(toIP),
		Port: toPort,
	}
	r := clientREST.New(to)
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
