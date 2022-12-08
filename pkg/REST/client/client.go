package clientREST

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	controller "github.com/chumvan/confdb/controllers"
	model "github.com/chumvan/confdb/models"
	"github.com/chumvan/go-sip-ua/pkg/utils"
	"github.com/chumvan/sipRestServer/pkg/topic"
	"github.com/ghettovoice/gosip/log"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.InfoLevel, "REST-Client", nil)
}

type ClientREST struct {
	ToIPAddr *net.TCPAddr
	ToURL    *url.URL

	FactoryAddr *net.UDPAddr

	client http.Client
}

type ResponseTopicInfo struct {
	Data struct {
		ConfUri   string `json:"confUri"`
		TopicIP   string `json:"topicIP"`
		TopicPort string `json:"topicPort"`
	} `json:"data"`
}

func New(restServerAddr *net.TCPAddr, factoryAddr *net.UDPAddr) (cr *ClientREST) {
	cr = &ClientREST{
		ToIPAddr:    restServerAddr,
		FactoryAddr: factoryAddr,
	}
	toUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", restServerAddr.IP.String(), restServerAddr.Port))
	if err != nil {
		return nil
	}
	cr.ToURL = toUrl
	cr.client = http.Client{Timeout: time.Duration(3) * time.Second}
	logger.Infof("created REST-client: \n\t sending to: %s", cr.ToIPAddr.String())
	return cr
}

func (cr *ClientREST) CreateTopic(meta topic.TopicMeta, chanInfo chan topic.TopicInfo) {
	// logger.Infof("received topic: %v", meta.Topic)
	confUri := fmt.Sprintf("sip:%s@%s", meta.Topic, cr.FactoryAddr)
	// logger.Infof("confUri: %s", confUri)
	// logger.Infof("creator: %s", meta.CreatorSipUri)

	confInfo := controller.ConfInfoInput{
		ConfUri: confUri,
		Subject: meta.Topic,
		Creator: model.User{
			EntityUrl: meta.CreatorSipUri.String(),
			Role:      "publisher",
		},
	}

	input, err := json.Marshal(confInfo)
	if err != nil {
		logger.Error(err)
	}

	url := cr.ToURL.String() + "/api/v1/confInfos"

	resp, err := cr.client.Post(url, "application/json", bytes.NewBuffer(input))
	if err != nil {
		logger.Error(err)
		return
	}
	// logger.Infof("full response: %s\n", *resp)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("ReadAll: \n", err)
	}
	// logger.Infof("body: %s\n", string(body))

	// print to debug
	// b, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	logger.Error(err)
	// }
	// logger.Infof("response: %s", string(b))

	// parse to send to topic info channel
	var data ResponseTopicInfo
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Error(err)
	}
	forwarderIP := os.Getenv("FORWARDER_IP")
	forwarderRtpInPort := os.Getenv("FORWARDER_RTP_IN_PORT")

	topicInfo := topic.TopicInfo{
		ConfUri:   data.Data.ConfUri,
		TopicIP:   forwarderIP,
		TopicPort: forwarderRtpInPort,
	}

	chanInfo <- topicInfo

}
