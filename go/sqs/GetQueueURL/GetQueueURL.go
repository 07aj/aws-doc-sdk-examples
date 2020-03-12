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
// snippet-start:[sqs.go.get_queue_url]
package main

// snippet-start:[sqs.go.get_queue_url.imports]
import (
    "flag"
    "fmt"

    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
)
// snippet-end:[sqs.go.get_queue_url.imports]

// GetQueueURL gets the URL of an Amazon SQS queue
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     queueName is the name of the queue
// Output:
//     If success, the URL of the queue and nil
//     Otherwise, an empty string and an error from the call to
func GetQueueURL(sess *session.Session, queueName *string) (string, error) {
    // Create an SQS service client
    // snippet-start:[sqs.go.get_queue_url.call]
    svc := sqs.New(sess)

    result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
        QueueName: queueName,
    })
    // snippet-end:[sqs.go.get_queue_url.call]
    if err != nil {
        return "", err
    }

    return *result.QueueUrl, nil
}

func main() {
    // snippet-start:[sqs.go.get_queue_url.args]
    queueName := flag.String("n", "", "The name of the queue")
    flag.Parse()

    if *queueName == "" {
        fmt.Println("You must supply a queue name (-n QUEUE-NAME")
        return
    }
    // snippet-end:[sqs.go.get_queue_url.args]

    // Create a session that gets credential values from ~/.aws/credentials
    // and the default region from ~/.aws/config
    // snippet-start:[sqs.go.get_queue_url.sess]
    sess := session.Must(session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    }))
    // snippet-end:[sqs.go.get_queue_url.sess]

    url, err := GetQueueURL(sess, queueName)
    if err != nil {
        fmt.Println("Got an error getting the queue URL:")
        fmt.Println(err)
        return
    }

    fmt.Println("URL for queue " + *queueName + ": " + url)
}
// snippet-end:[sqs.go.get_queue_url]
