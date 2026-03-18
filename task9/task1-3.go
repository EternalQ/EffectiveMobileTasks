package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

func task1() {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll

	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		panic(err)
	}
	defer admin.Close()

	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
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

	cons, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		log.Fatalln("consumer create err: ", err)
	}
	defer cons.Close()

	pc, err := cons.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalln("pcons create err: ", err)
	}
	defer pc.Close()

	go func() {
		for err := range pc.Errors() {
			log.Println("send err: ", err)
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := range 10 {
		msg := fmt.Sprintf("msg #%d", i)
		sendMsg(prod, 1, msg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	for range 3 {
		go func(ctx context.Context, wg *sync.WaitGroup) {
		loop:
			for {
				select {
				case <-ctx.Done():
					break loop
				case msg := <-pc.Messages():
					fmt.Println("readed: ", string(msg.Value))
					time.Sleep(400 * time.Millisecond) // working

					wg.Done()
				}
			}
		}(ctx, wg)
	}
	wg.Wait()
	cancel()
}
