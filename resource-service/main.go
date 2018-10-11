package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

var (
	brokerAddress = fmt.Sprintf("%s:%s", os.Getenv("KAFKA_ADDRESS"), os.Getenv("KAFKA_PORT"))
)

func main() {
	fmt.Println("Running")

	messageBrokerReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerAddress},
		Topic:     "github",
		Partition: 0,
	})

	// Only read new messages (according to Kafka)
	err := messageBrokerReader.SetOffset(-2)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		message, err := messageBrokerReader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(err)
		} else {
			eventType := string(message.Key)
			if eventType == "installation_repositories" {
				err := processInstallationRepositoriesMessage(message.Value)
				if err != nil {
					fmt.Println(err)
					break
				}
			}
		}
	}

	err = messageBrokerReader.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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

			err = publishEvent([]byte("repo_added"), repoUpdatedMessage)
			if err != nil {
				return err
			}
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

			err = publishEvent([]byte("repo_removed"), repoUpdatedMessage)
			if err != nil {
				return err
			}
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
	messageBusWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   "coalesce",
	})

	fmt.Println("Sending", string(key), string(value))
	err := messageBusWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   key,
			Value: value,
		},
	)
	if err != nil {
		return err
	}

	return messageBusWriter.Close()
}
