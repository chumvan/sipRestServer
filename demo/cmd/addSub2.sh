curl --location --request PATCH 'http://localhost:8080/api/v1/topicMode/confInfos/amazingTopic/' \
--header 'Content-Type: application/json' \
--data-raw '{
    "EntityURL": "sip:subscriber2@10.5.0.5:5072",
    "Role": "subscriber",
    "PortRTP": 6422
}' | jq