package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"github.com/jszwedko/go-circleci"
	"github.com/segmentio/kafka-go"
)

var (
	client *circleci.Client
	db     *sql.DB

	brokerAddress = fmt.Sprintf("%s:%s", os.Getenv("KAFKA_ADDRESS"), os.Getenv("KAFKA_PORT"))
)

func main() {
	fmt.Println("Running")

	client = &circleci.Client{Token: os.Getenv("CIRCLECI_TOKEN")}

	messageBrokerReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerAddress},
		Topic:     "coalesce",
		Partition: 0,
	})

	dbConnect()

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
			if eventType == "repo_added" {
				err := processRepoAddedMessage(message.Value)
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

func dbConnect() {
	var err error
	connStr := "postgres://docker:docker@circleci-db:5000/coalesce_circleci?sslmode=disable"
	fmt.Println("Connecting to DB")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func processRepoAddedMessage(message []byte) error {
	var repoUpdatedMessage RepoUpdatedMessage
	err := json.Unmarshal(message, &repoUpdatedMessage)
	if err != nil {
		return err
	}

	fmt.Printf("Repository %q added by %s, fetching some info from CircleCI\n", repoUpdatedMessage.RepoName, repoUpdatedMessage.User)

	repoNameParts := strings.Split(repoUpdatedMessage.RepoName, "/")

	// Follow the project first because CircleCI's API is found wanting
	_, err = client.FollowProject(repoNameParts[0], repoNameParts[1])
	if err != nil {
		return err
	}

	project, err := client.GetProject(repoNameParts[0], repoNameParts[1])
	if err != nil {
		return err
	}

	if project == nil {
		fmt.Printf("Repository %q is not set up in CircleCI!\n", repoUpdatedMessage.RepoName)
	} else {
		fmt.Printf("Repository %q is set up in CircleCI, neato!\n", repoUpdatedMessage.RepoName)

		slackWebhookURL := ""
		slackWebhookID := ""
		if project.SlackWebhookURL != nil {
			slackWebhookURL = *project.SlackWebhookURL

			slackWebhookURLParts := strings.Split(slackWebhookURL, "/")
			if len(slackWebhookURLParts) > 0 {
				slackWebhookID = slackWebhookURLParts[len(slackWebhookURLParts)-2]
			}
		}

		db.QueryRow(fmt.Sprintf(`INSERT INTO circleci(repo_full_name, slack_webhook_url, slack_webhook_id) VALUES('%s', '%s', '%s')`, repoUpdatedMessage.RepoName, slackWebhookURL, slackWebhookID))
		if err != nil {
			return err
		}
	}

	return nil
}
