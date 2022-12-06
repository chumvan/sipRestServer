package server

import (
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
	f := conffactory.NewConfFactory()
	if f == nil {
		return nil
	}
	s := &Server{
		Factory: f,
	}
	return s
}

func (s *Server) Shutdown() {
	s.Factory.Shutdown()
	s.SIP.Shutdown()
}
