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
		Topic:     "github",
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
		if eventType == "installation_repositories" {
			err := processInstallationRepositoriesMessage(message.Value)
			if err != nil {
				fmt.Println(err)
				break
			}
		}

	}

	messageBrokerReader.Close()
}

func processInstallationRepositoriesMessage(message []byte) error {
	var githubInstallationRepositoriesMessage GithubInstallationRepositoriesMessage
	err := json.Unmarshal(message, &githubInstallationRepositoriesMessage)
	if err != nil {
		return err
	}

	switch githubInstallationRepositoriesMessage.Action {
	case "added":
		fmt.Printf("Repositories added by %s, save it to the database and process\n", githubInstallationRepositoriesMessage.Sender.Login)
		for _, repo := range githubInstallationRepositoriesMessage.RepositoriesAdded {
			repoUpdatedMessage, err := createRepoUpdatedMessage(
				repo.FullName,
				githubInstallationRepositoriesMessage.Sender.Login,
			)

			if err != nil {
				return err
			}

			publishEvent([]byte("repo_added"), repoUpdatedMessage)
		}

	case "removed":
		fmt.Printf("Repositories removed by %s, delete all the things\n", githubInstallationRepositoriesMessage.Sender.Login)
		for _, repo := range githubInstallationRepositoriesMessage.RepositoriesRemoved {
			repoUpdatedMessage, err := createRepoUpdatedMessage(
				repo.FullName,
				githubInstallationRepositoriesMessage.Sender.Login,
			)

			if err != nil {
				return err
			}

			publishEvent([]byte("repo_removed"), repoUpdatedMessage)
		}
	}

	return nil
}

func createRepoUpdatedMessage(repoName, user string) ([]byte, error) {
	repoUpdatedMessage := RepoUpdatedMessage{
		RepoName: repoName,
		User:     user,
	}

	return json.Marshal(repoUpdatedMessage)
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
