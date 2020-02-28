/*
   Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

   This file is licensed under the Apache License, Version 2.0 (the "License").
   You may not use this file except in compliance with the License. A copy of
   the License is located at

    http://aws.amazon.com/apache2.0/

   This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
   CONDITIONS OF ANY KIND, either express or implied. See the License for the
   specific language governing permissions and limitations under the License.
*/

package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// Config defines a set of configuration values
type Config struct {
	QueueName   string `json:"QueueName"`
	DlQueueName string `json:"DlQueueName"`
}

// configFile defines the name of the file containing configuration values
var configFileName = "config.json"

// globalConfig contains the configuration values
var globalConfig Config

func populateConfiguration(t *testing.T) error {
	// Get configuration from config.json

	// Get entire file as a JSON string
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	// Convert []byte to string
	text := string(content)

	// Marshall JSON string in text into global struct
	err = json.Unmarshal([]byte(text), &globalConfig)
	if err != nil {
		return err
	}

	id := uuid.New()

	if globalConfig.QueueName == "" {
		// Create unique, random queue name
		globalConfig.QueueName = "myqueue-" + id.String()
	}

	if globalConfig.DlQueueName == "" {
		// Create unique, random queue name
		globalConfig.DlQueueName = "mydlqueue-" + id.String()
	}

	return nil
}

func createQueue(sess *session.Session, queueName string) (string, error) {
	// Create a SQS service client
	svc := sqs.New(sess)

	result, err := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]*string{
			"DelaySeconds":           aws.String("60"),
			"MessageRetentionPeriod": aws.String("86400"),
		},
	})
	if err != nil {
		return "", err
	}

	return *result.QueueUrl, nil
}

func getQueueArn(queueURL string) (string, error) {
	parts := strings.Split(queueURL, "/")

	subParts := strings.Split(parts[2], ".")

	return "arn:aws:" + subParts[0] + ":" + subParts[1] + ":" + parts[3] + ":" + parts[4], nil
}

func deleteQueue(sess *session.Session, queueURL string) error {
	// Create a SQS service client
	svc := sqs.New(sess)

	_, err := svc.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		return err
	}

	return nil
}

func TestQueue(t *testing.T) {
	err := populateConfiguration(t)
	if err != nil {
		t.Fatal(err)
	}

	// Create a session using credentials from ~/.aws/credentials
	// and the region from ~/.aws/config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	createURL, err := createQueue(sess, globalConfig.QueueName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Got URL " + createURL + " for queue " + globalConfig.QueueName)

	// dlURL, dlARN, err := createDeadLetterQueue(t, sess, globalConfig.DlQueueName)
	dlURL, err := createQueue(sess, globalConfig.DlQueueName)
	if err != nil {
		t.Fatal(err)
	}

	dlARN, err := getQueueArn(dlURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Created dead-letter queue " + globalConfig.DlQueueName)

	err = ConfigureDeadLetterQueue(sess, dlARN, createURL)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteQueue(sess, createURL)
	if err != nil {
		t.Log("You'll have to delete queue " + globalConfig.QueueName + " yourself")
		t.Fatal(err)
	}

	err = deleteQueue(sess, dlURL)
	if err != nil {
		t.Log("You'll have to delete queue " + globalConfig.DlQueueName + " yourself")
		t.Fatal(err)
	}

	t.Log("Deleted queue " + globalConfig.QueueName)
	t.Log("Deleted queue " + globalConfig.DlQueueName)
}
