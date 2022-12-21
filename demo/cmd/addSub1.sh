curl --location --request PATCH 'http://localhost:8080/api/v1/topicMode/confInfos/amazingTopic/' \
--header 'Content-Type: application/json' \
--data-raw '{
    "EntityURL": "sip:subscriber1@10.5.0.4:5071",
    "Role": "subscriber",
    "PortRTP": 6421
}' | jq