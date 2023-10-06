package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Checkout struct {
	OrderID string `json:"order_id"`
	Amount  int    `json:"amount"`
}

func main() {
	forever := make(chan bool)
	go runConsumer()
	time.Sleep(time.Second * 10)
	go runProducer()
	time.Sleep(time.Second * 10)
	go runProducer()
	<-forever

}

func runConsumer() {
	conn, err := amqp091.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	q, err := channel.QueueDeclare(
		"test", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	if err != nil {
		panic(err)
	}

	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		panic(err)
	}
	fmt.Println("Waiting for messages...")
	for msg := range msgs {
		msg.Ack(false)
		execute(msg)
	}
}

func execute(msg amqp091.Delivery) {
	var data Checkout
	if err := json.Unmarshal(msg.Body, &data); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("executing %v\n", string(data.OrderID))
	time.Sleep(time.Millisecond * 500)
}

func runProducer() {
	conn, err := amqp091.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	q, err := channel.QueueDeclare(
		"test", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	if err != nil {
		panic(err)
	}
	// return

	length := 10

	for i := 0; i < length; i++ {

		data := Checkout{
			OrderID: fmt.Sprintf("order-%v", i),
			Amount:  i * 100,
		}
		dataJson, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		if err := channel.PublishWithContext(context.Background(), "", q.Name, false, false, amqp091.Publishing{
			ContentType: "application/json",
			Body:        dataJson,
		}); err != nil {
			fmt.Println(err.Error())
		}

	}

	fmt.Printf("published %v messages\n", length)
}
