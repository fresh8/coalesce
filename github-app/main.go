package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/segmentio/kafka-go"
)

var (
	brokerAddress = fmt.Sprintf("%s:%s", os.Getenv("KAFKA_ADDRESS"), os.Getenv("KAFKA_PORT"))
)

func main() {
	fmt.Println("Running")
	http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		fmt.Println("POST received")

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Println(err)
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		}

		messageBusWriter := kafka.NewWriter(kafka.WriterConfig{
			Brokers: []string{brokerAddress},
			Topic:   "github",
		})

		err = messageBusWriter.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(request.Header.Get("X-Github-Event")),
				Value: body,
			},
		)
		if err != nil {
			fmt.Println(err)
		}

		err = messageBusWriter.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Fprintf(responseWriter, "Received")
	})

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
