package mock

import (
	"net"
	"time"

	"github.com/pixelbender/go-sdp/sdp"
)

var (
	host   = "127.0.0.1"
	Offer  *sdp.Session
	Answer *sdp.Session
)

func BuildInviteWithTopic(addr *net.UDPAddr, topic string) string {
	host := addr.IP.String()
	port := addr.Port

	sdp := &sdp.Session{
		Origin: &sdp.Origin{
			Username:       "-",
			Address:        host,
			SessionID:      time.Now().UnixNano() / 1e6,
			SessionVersion: time.Now().UnixNano() / 1e6,
		},
		Timing: &sdp.Timing{Start: time.Time{}, Stop: time.Time{}},
		//Name: "Example",
		Connection: &sdp.Connection{
			Address: host,
		},
		//Bandwidth: []*sdp.Bandwidth{{Type: "AS", Value: 117}},
		Media: []*sdp.Media{
			{
				//Bandwidth: []*sdp.Bandwidth{{Type: "TIAS", Value: 96000}},
				Connection: []*sdp.Connection{{Address: host}},
				Mode:       sdp.SendRecv,
				Type:       "text",
				Port:       port,
				Proto:      "RTP/AVP",
				Format: []*sdp.Format{
					{Payload: 100, Name: "t140", ClockRate: 1000},
				},
			},
		},
		Attributes: []*sdp.Attr{
			sdp.NewAttr("topic", topic),
		},
	}

	return sdp.String()
}

func BuildLocalSdp(host string, port int) string {
	sdp := &sdp.Session{
		Origin: &sdp.Origin{
			Username:       "-",
			Address:        host,
			SessionID:      time.Now().UnixNano() / 1e6,
			SessionVersion: time.Now().UnixNano() / 1e6,
		},
		Timing: &sdp.Timing{Start: time.Time{}, Stop: time.Time{}},
		//Name: "Example",
		Connection: &sdp.Connection{
			Address: host,
		},
		//Bandwidth: []*sdp.Bandwidth{{Type: "AS", Value: 117}},
		Media: []*sdp.Media{
			{
				//Bandwidth: []*sdp.Bandwidth{{Type: "TIAS", Value: 96000}},
				Connection: []*sdp.Connection{{Address: host}},
				Mode:       sdp.SendRecv,
				Type:       "audio",
				Port:       port,
				Proto:      "RTP/AVP",
				Format: []*sdp.Format{
					{Payload: 0, Name: "PCMU", ClockRate: 8000},
					{Payload: 8, Name: "PCMA", ClockRate: 8000},
					//{Payload: 18, Name: "G729", ClockRate: 8000, Params: []string{"annexb=yes"}},
					{Payload: 106, Name: "telephone-event", ClockRate: 8000, Params: []string{"0-16"}},
				},
			},
		},
	}
	return sdp.String()
}

func GetRemoteIpPort(sdp *sdp.Session) (string, int) {
	return sdp.Connection.Address, sdp.Media[0].Port
}
