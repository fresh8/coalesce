package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jszwedko/go-circleci"
	"github.com/segmentio/kafka-go"
)

var (
	client *circleci.Client
)

func main() {
	fmt.Println("Running")

	client = &circleci.Client{Token: os.Getenv("CIRCLECI_TOKEN")}

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
	var repoUpdatedMessage RepoUpdatedMessage
	err := json.Unmarshal(message, &repoUpdatedMessage)
	if err != nil {
		return err
	}

	fmt.Printf("Repository %q added by %s, fetching some info from CircleCI\n", repoUpdatedMessage.RepoName, repoUpdatedMessage.User)

	repoNameParts := strings.Split(repoUpdatedMessage.RepoName, "/")

	project, err := client.GetProject(repoNameParts[0], repoNameParts[1])
	if err != nil {
		return err
	}

	if project == nil {
		fmt.Printf("Repository %q is not set up in CircleCI!\n", repoUpdatedMessage.RepoName)
	} else {
		fmt.Printf("Repository %q is set up in CircleCI, neato!\n", repoUpdatedMessage.RepoName)
	}

	return nil
}

func publishEvent(key, value []byte) error {
	brokerAddress := fmt.Sprintf("%s:9092", os.Getenv("KAFKA_ADDRESS"))
	messageBusWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   "dredd",
	})

	fmt.Println("Sending", string(key), string(value))
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
