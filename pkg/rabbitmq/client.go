package rabbitmq

import (
	"github.com/streadway/amqp"
)

type Conn struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewConn(queueUrl string) (*Conn, error) {
	conn, err := amqp.Dial(queueUrl)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	channel.Qos(1, 0, false) // Set QoS to ensure fair dispatch

	return &Conn{
		Conn:    conn,
		Channel: channel,
	}, nil
}

func (c *Conn) Close() {
	if c.Channel != nil {
		c.Channel.Close()
	}
	if c.Conn != nil {
		c.Conn.Close()
	}
}
