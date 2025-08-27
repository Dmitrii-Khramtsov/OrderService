package kafka

import "errors"

var (
	ErrKafkaConnectionFailed = errors.New("failed to connect to kafka")
	ErrKafkaMessageFetch     = errors.New("failed to fetch kafka message")
	ErrKafkaMessageCommit    = errors.New("failed to commit kafka message")
	ErrKafkaMessageDecode    = errors.New("failed to decode kafka message")
	ErrKafkaOrderSave        = errors.New("failed to save order from kafka message")
)
