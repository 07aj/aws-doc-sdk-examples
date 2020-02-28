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
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// ConfigureDeadLetterQueue configures an Amazon SQS queue for messages that could not be delivered to another queue
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     deadLetterQueueARN is the ARN of the dead-letter queue
//     queueURL is the URL of the queue that did not get messages
// Output:
//     If success, the URL of the queue and nil
//     Otherwise, an empty string and an error from the call to json.Marshal or SetQueueAttributes
func ConfigureDeadLetterQueue(sess *session.Session, deadLetterQueueARN string, queueURL string) error {
	// Create a SQS service client
	svc := sqs.New(sess)

	// Our redrive policy for our queue
	policy := map[string]string{
		"deadLetterTargetArn": deadLetterQueueARN,
		"maxReceiveCount":     "10",
	}

	// Marshal policy for SetQueueAttributes
	b, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	_, err = svc.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		QueueUrl: aws.String(queueURL),
		Attributes: map[string]*string{
			sqs.QueueAttributeNameRedrivePolicy: aws.String(string(b)),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	queueURLPtr := flag.String("u", "", "The URL of the queue")
	dlQueueARN := flag.String("d", "", "The ARN of the dead-letter queue")
	flag.Parse()

	if *queueURLPtr == "" || *dlQueueARN == "" {
		fmt.Println("You must supply the URL of the queue (-u QUEUE-URL) and the ARN of the dead-letter queue (-d QUEUE-ARN)")
		return
	}

	// Create a session that get credential values from ~/.aws/credentials
	// and the default region from ~/.aws/config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	err := ConfigureDeadLetterQueue(sess, *dlQueueARN, *queueURLPtr)
	if err != nil {
		fmt.Println("Got an error creating the dead-letter queue:")
		fmt.Println(err)
		return
	}

	fmt.Println("Created dead-letter queue")
}
