package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

func main() {
	fmt.Println("Running")

	brokerAddress := fmt.Sprintf("%s:9092", os.Getenv("KAFKA_ADDRESS"))
	messageBrokerReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerAddress},
		Topic:     "dredd",
		Partition: 0,
	})

	// Only read new messages (according to Kafka)
	messageBrokerReader.SetOffset(-2)

	for {
		message, err := messageBrokerReader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(err)
			break
		}

		eventType := string(message.Key)
		if eventType == "repo_added" {
			err := processRepoAddedMessage(message.Value)
			if err != nil {
				fmt.Println(err)
				break
			}
		}

	}

	messageBrokerReader.Close()
}

func processRepoAddedMessage(message []byte) error {
	var repoAddedMessage RepoAddedMessage
	err := json.Unmarshal(message, &repoAddedMessage)
	if err != nil {
		return err
	}

	fmt.Printf("Repository %q added by %s, fetching some info from CircleCI\n", repoAddedMessage.RepoName, repoAddedMessage.User)

	return nil
}

func publishEvent(key, value []byte) error {
	brokerAddress := fmt.Sprintf("%s:9092", os.Getenv("KAFKA_ADDRESS"))
	messageBusWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   "dredd",
	})

	err := messageBusWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   key,
			Value: value,
		},
	)
	if err != nil {
		fmt.Println(err)
	}

	messageBusWriter.Close()

	return nil
}
