package main

import (
	"fmt"

	"github.com/IBM/sarama"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "test-topic"
)

func main() {
	// task1()
	task4()
}

func sendMsg(prod sarama.AsyncProducer, part int, str string) {
	key := fmt.Sprintf("part-%d", part)
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(str),
		// Partition: int32(part), //task 5
		Key: sarama.StringEncoder(key),
	}
	prod.Input() <- msg
}
