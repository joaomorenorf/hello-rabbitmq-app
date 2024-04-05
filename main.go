package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

const version = "1.0.0"

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connectToRabbitMQ() *amqp.Connection {
	// Load database credentials from environment variables
	pass := ""
	user, exist := os.LookupEnv("RABBITMQ_DEFAULT_USER")
	if !exist {
		user = "guest"
		pass = "guest"
		log.Print("Failed to retrieve optional RABBITMQ_DEFAULT_USER variable, using guest:guest")
	} else {
		pass, exist = os.LookupEnv("RABBITMQ_DEFAULT_PASS")
		if !exist {
			log.Fatalf("Failed to retrieve RABBITMQ_DEFAULT_PASS variable")
		}
	}
	server, exist := os.LookupEnv("RABBITMQ_SERVER")
	if !exist {
		log.Fatalf("Failed to retrieve RABBITMQ_SERVER variable")
	}
	vhost, exist := os.LookupEnv("RABBITMQ_DEFAULT_VHOST")
	if !exist {
		log.Print("Failed to retrieve optional RABBITMQ_DEFAULT_VHOST variable, using /")
	}

	// Construct connection string
	u := &url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(user, pass),
		Host:   server,
		Path:   vhost,
	}
	conn, err := amqp.Dial(u.String())
	failOnError(err, "Failed to connect to RabbitMQ")
	return conn
}

func sendMessage(ch *amqp.Channel, q amqp.Queue, msg string) {
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	failOnError(err, "Failed to publish a message")
}

func receiveMessage(msgs <-chan amqp.Delivery) *string {
	var body string
	select {
	case x, ok := <-msgs:
		if ok {
			body = string(x.Body)
			return &body
		} else {
			log.Fatalf("Channel closed!\n")
		}
	default:
		return nil
	}
	return nil
}

func setupRoutes(ch *amqp.Channel, q amqp.Queue, msgs <-chan amqp.Delivery) {

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().Format(time.RFC3339)
		msg := fmt.Sprintf("Hello, world!\nVersion: %s\nSent at: %s\nSender: %s\n", version, now, r.RemoteAddr)
		sendMessage(ch, q, msg)
		fmt.Fprintf(w, "%s", msg)
		log.Printf("Message queued\tSender: %s\n", r.RemoteAddr)
	})

	http.HandleFunc("/consume", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().Format(time.RFC3339)
		msg := receiveMessage(msgs)
		if msg == nil {
			fmt.Fprintf(w, "No messages queued\tRequested by: %s\n", r.RemoteAddr)
			log.Printf("No messages queued\tRequested by: %s\n", r.RemoteAddr)
		} else {
			fmt.Fprintf(w, "%s\nConsumed at: %s\nConsumed by: %s\n", *msg, now, r.RemoteAddr)
			date := strings.Split(*msg, "Sent at: ")[1]
			date = strings.Split(date, "\nSender: ")[0]
			log.Printf("Message consumed\tSent at: %s\tRequested by: %s\n", date, r.RemoteAddr)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8008"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func startQueue(conn *amqp.Connection) (*amqp.Channel, amqp.Queue, <-chan amqp.Delivery) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	return ch, q, msgs
}

func main() {
	conn := connectToRabbitMQ()
	defer conn.Close()
	ch, q, msgs := startQueue(conn)
	defer ch.Close()
	setupRoutes(ch, q, msgs)
}
