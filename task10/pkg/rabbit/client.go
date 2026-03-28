package rabbit

import (
	"fmt"
	"log"

	rmq "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	conn *rmq.Connection
	ch   *rmq.Channel
	exch string
}

func NewClient(host, exchange string) *RabbitClient {
	str := fmt.Sprintf("amqp://guest:guest@%s/", host)
	conn, err := rmq.Dial(str)
	if err != nil {
		log.Fatalf("Dial err: %v\n", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Conn err: %v\n", err)
	}

	if err := ch.ExchangeDeclare(
		exchange,
		rmq.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("Exch err: %v\n", err)
	}

	return &RabbitClient{conn, ch, exchange}
}

func (c *RabbitClient) Close() {
	c.conn.Close()
	c.ch.Close()
}

func (c *RabbitClient) Pub(msg string) error {
	return c.ch.Publish(c.exch, "", false, false, rmq.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})
}

func (c *RabbitClient) Sub() (<-chan rmq.Delivery, error) {
	q, err := c.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	if err := c.ch.QueueBind(
		q.Name,
		"",
		c.exch,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	return c.ch.Consume(q.Name, "", true, false, false, false, nil)
}
