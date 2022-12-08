package topic

import "github.com/ghettovoice/gosip/sip"

type TopicMeta struct {
	Topic         string
	CreatorSipUri sip.Uri
}

type TopicInfo struct {
	ConfUri   string
	TopicIP   string
	TopicPort string
}
