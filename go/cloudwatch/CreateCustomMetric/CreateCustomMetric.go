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
// snippet-start: [cloudwatch.go.create_custom_metric]
package main

// snippet-start: [cloudwatch.go.create_custom_metric.import]
import (
    "flag"
    "fmt"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/cloudwatch"
)
// snippet-end: [cloudwatch.go.create_custom_metric.import]

// CreateCustomMetric creates a new metric in a namespace
// Inputs:
//     namespace is the metric namespace
//     metricName is the name of the metric
//     unit is what the value represents
//     value is the value of the metric unit
//     dimensionName is the name of the dimension
//     dimensionValue is the value of the dimensionName
// Output:
//     If success, nil
//     Otherwise, and error from a call to PutMetricData
func CreateCustomMetric(sess *session.Session, namespace string, metricName string, unit string, value float64, dimensionName string, dimensionValue string) error {
    // snippet-start: [cloudwatch.go.create_custom_metric.client]
    // Create new cloudwatch client
    svc := cloudwatch.New(sess)
    // snippet-end: [cloudwatch.go.create_custom_metric.client]

    // snippet-start: [cloudwatch.go.create_custom_metric.call]
    _, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
        Namespace: aws.String(namespace),
        MetricData: []*cloudwatch.MetricDatum{
            &cloudwatch.MetricDatum{
                MetricName: aws.String(metricName),
                Unit:       aws.String(unit),
                Value:      aws.Float64(value),
                Dimensions: []*cloudwatch.Dimension{
                    &cloudwatch.Dimension{
                        Name:  aws.String(dimensionName),
                        Value: aws.String(dimensionValue),
                    },
                },
            },
        },
    })
    // snippet-end: [cloudwatch.go.create_custom_metric.call]
    if err != nil {
        return err
    }

    return nil
}

func main() {
    namespacePtr := flag.String("n", "", "The namespace for the metric")
    metricNamePtr := flag.String("m", "", "The name of the metric")
    unitPtr := flag.String("u", "", "The units for the metric")
    valuePtr := flag.Float64("v", 0.0, "The value of the units")
    dimensionNamePtr := flag.String("dn", "", "The name of the dimension")
    dimensionValuePtr := flag.String("dv", "", "The value of the dimension")
    flag.Parse()
    namespace := *namespacePtr
    metricName := *metricNamePtr
    unit := *unitPtr
    value := *valuePtr
    dimensionName := *dimensionNamePtr
    dimensionValue := *dimensionValuePtr

    // Initialize a session that the SDK uses to load
    // credentials from the shared credentials file ~/.aws/credentials
    // and configuration from the shared configuration file ~/.aws/config.
    sess := session.Must(session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    }))

    err := CreateCustomMetric(sess, namespace, metricName, unit, value, dimensionName, dimensionValue)
    if err != nil {
        fmt.Println()
    }
}
// snippet-end: [cloudwatch.go.create_custom_metric]