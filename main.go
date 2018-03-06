package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

var (
	uri         = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	insecureTls = flag.Bool("insecure-tls", false, "Insecure TLS mode: don't check certificates")
	queue       = flag.String("queue", "", "AMQP queue name")
	ack         = flag.Bool("ack", false, "Acknowledge messages")
	maxMessages = flag.Uint("max-messages", 1000, "Maximum number of messages to dump")
	messages    = make([][]byte, uint(*maxMessages)) // [flag.Uint("max-messages", 1000, "Maximum number of messages to dump")] []byte
)

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unused command line arguments detected.\n")
		flag.Usage()
		os.Exit(2)
	}
	err := DumpMessagesFromQueue(*uri, *queue, *maxMessages/*, *outputDir*/)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	messagesJson, err := json.Marshal(messages)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "%s\n", messagesJson)
}

func dial(amqpURI string) (*amqp.Connection, error) {
	if *insecureTls && strings.HasPrefix(amqpURI, "amqps://") {
		tlsConfig := new(tls.Config)
		tlsConfig.InsecureSkipVerify = true
		conn, err := amqp.DialTLS(amqpURI, tlsConfig)
		return conn, err
	}
	conn, err := amqp.Dial(amqpURI)
	return conn, err
}

func DumpMessagesFromQueue(amqpURI string, queueName string, maxMessages uint) error {
	if queueName == "" {
		return fmt.Errorf("Must supply queue name")
	}

	conn, err := dial(amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	defer func() {
		conn.Close()
	}()

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	for messagesReceived := uint(0); messagesReceived < maxMessages; messagesReceived++ {
		msg, ok, err := channel.Get(queueName,
			*ack, // autoAck
		)
		if err != nil {
			return fmt.Errorf("Queue get: %s", err)
		}

		if !ok {
			break
		}

		messages[messagesReceived] = msg.Body

		if err != nil {
			return fmt.Errorf("Save message: %s", err)
		}
	}

	return nil
}