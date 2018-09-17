package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/segmentio/kafka-go"
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

		brokerAddress := fmt.Sprintf("%s:9092", os.Getenv("KAFKA_ADDRESS"))
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

		messageBusWriter.Close()

		fmt.Fprintf(responseWriter, "Received")
	})

	http.ListenAndServe(":3000", nil)
}
