package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
)

type bHandler struct {
	limit   int
	counter atomic.Int32
	done    chan struct{}
}

func (h *bHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *bHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *bHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		time.Sleep(400 * time.Millisecond)
		fmt.Printf("readed: %s - %s\n", string(msg.Key), string(msg.Value))
		sess.MarkMessage(msg, "") // task9
		if h.counter.Add(1) == int32(h.limit) {
			close(h.done)
			return nil
		}
	}
	return nil
}

func task4() {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Offsets.AutoCommit.Enable = true // task 10 (default)

	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		panic(err)
	}
	defer admin.Close()

	admin.DeleteTopic(topic)
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
		log.Fatal("topic create error: ", err)
	}
	fmt.Println("topic created: ", topic)

	prod, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		log.Fatalln("producer create err: ", err)
	}
	defer prod.Close()

	group, err := sarama.NewConsumerGroup(brokers, "test-group", cfg)
	if err != nil {
		log.Fatal("consGroup create err: ", err)
	}

	limit := 10
	ctx, cancel := context.WithCancel(context.Background())
	handler := &bHandler{limit: limit, done: make(chan struct{})}

	for range 3 {
		go func(ctx context.Context) {
			for {
				if err := group.Consume(ctx, []string{topic}, handler); err != nil {
					log.Fatal("consume err: ", err)
				}
			}
		}(ctx)
	}

	for i := range limit {
		msg := fmt.Sprintf("msg #%d", i)
		sendMsg(prod, i%3, msg)
	}

	<-handler.done
	cancel()
}
