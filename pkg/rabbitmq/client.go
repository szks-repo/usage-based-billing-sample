package rabbitmq

import (
	"github.com/streadway/amqp"
)"

type Conn struct {
	conn *amqp.Connection
	channel *amqp.Channel
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
		conn:    conn,
		channel: channel,
	}, nil
}

func (c *Conn) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
