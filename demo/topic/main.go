package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	model "github.com/chumvan/confdb/models"
	router "github.com/chumvan/forwarder-rest-server/routers"
	utils "github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/chumvan/rtp/forwarder"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip/parser"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.DebugLevel, "Topicer", nil)
}

func main() {
	f := forwarder.Forwarder{}
	topicIP := os.Getenv("FORWARDER_IP")
	topicInPortStr := os.Getenv("FORWARDER_RTP_IN_PORT")
	topicInPort, err := strconv.Atoi(topicInPortStr)
	if err != nil {
		logger.Error(err)
	}
	inAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", topicIP, topicInPort))
	if err != nil {
		logger.Error(err)
	}
	topicOutPortStr := os.Getenv("FORWARDER_RTP_OUT_PORT")
	topicOutPort, err := strconv.Atoi(topicOutPortStr)
	if err != nil {
		logger.Error(err)
	}
	outAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", topicIP, topicOutPort))
	if err != nil {
		logger.Error(err)
	}
	err = f.SetupInConn(*inAddr, *outAddr)
	if err != nil {
		logger.Error(err)
	}

	addrChan := make(chan []net.UDPAddr, 1)
	defaultAddrs := getSubDefaultAddrs()
	addrChan <- defaultAddrs
	wg := new(sync.WaitGroup)

	// REST server to update address
	usersChan := make(chan []model.User, 1)
	r := router.SetupRouter(usersChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		restIP := os.Getenv("FORWADER_REST_PORT")
		r.Run(fmt.Sprintf(":%s", restIP))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for users := range usersChan {
			udpAddrs := parseUDPAddrsFromUsers(users)
			addrChan <- udpAddrs
		}
	}()

	// remote address Updaters
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.UpdateRemoteAddrs(addrChan)
	}()

	// UDP packets forwarder
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := f.Forward(); err != nil {
			panic(err)
		}
	}()

	wg.Wait()

}

func parseUDPAddrsFromUsers(users []model.User) (udpAddrs []net.UDPAddr) {
	fmt.Println("received users")
	for _, u := range users {
		fmt.Printf("full string: %v\n", u)
		uri, err := parser.ParseSipUri(u.EntityUrl)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("user: %s\n", uri.User())
		host, portStr, err := net.SplitHostPort(uri.Host())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("host: %s\n", host)
		fmt.Printf("port: %s\n", portStr)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println(err)
		}
		userAddr := net.UDPAddr{
			IP:   net.ParseIP(host),
			Port: port,
		}
		udpAddrs = append(udpAddrs, userAddr)
	}
	return
}

func getSubDefaultAddrs() []net.UDPAddr {
	var addrs []net.UDPAddr
	for _, v := range []string{"1", "2", "3"} {
		ip := os.Getenv(fmt.Sprintf("SUBSCRIBER%s_IP", v))
		portStr := os.Getenv(fmt.Sprintf("SUBSCRIBER%s_RTP_PORT", v))
		port, err := strconv.Atoi(portStr)
		if err != nil {
			logger.Error(err)
		}
		addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", ip, port))
		if err != nil {
			logger.Error(err)
		}
		addrs = append(addrs, *addr)
	}
	return addrs
}
