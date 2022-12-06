package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chumvan/sipRestServer/src/server"
)

func main() {
	s := server.New()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	s.Shutdown()
}
