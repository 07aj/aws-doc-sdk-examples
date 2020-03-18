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
    "errors"
    "fmt"
    "io/ioutil"
    "testing"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/cloudwatch"
    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/google/uuid"
)

type config struct {
    InstanceName string `json:"InstanceName"`
    InstanceID   string `json:"InstanceID"`
    AlarmName    string `json:"AlarmName"`
}

var globalConfig config

var configFileName = "config.json"

func getEc2Info(sess *session.Session, t *testing.T) (string, string, error) {
    svc := ec2.New(sess)

    result, err := svc.DescribeInstances(nil)
    if err != nil {
        return "", "", err
    }

    // Return name and ID of the first EC2 instance with both
    for _, reservation := range result.Reservations {
        for _, instance := range reservation.Instances {
            id := instance.InstanceId
            name := ""

            // Name is stashed in a tag
            for _, tag := range instance.Tags {
                if *tag.Key == "Name" {
                    name = *tag.Value
                    break
                }
            }

            if *id != "" && name != "" {
                return *id, name, nil
            }
        }
    }

    return "", "", errors.New("No EC2 instance found with name and ID")
}

func populateConfig(sess *session.Session, t *testing.T) error {
    // Get and store configuration values
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

    if globalConfig.InstanceName == "" || globalConfig.InstanceID == "" {
        // Get EC2 instance name and ID
        globalConfig.InstanceName, globalConfig.InstanceID, err = getEc2Info(sess, t)
        if err != nil {
            return err
        }
    }

    if globalConfig.AlarmName == "" {
        // Create random alarm name
        id := uuid.New()
        globalConfig.AlarmName = "Alarm70-" + id.String()
    }

    t.Log("Instance Name: " + globalConfig.InstanceName)
    t.Log("Instance ID:   " + globalConfig.InstanceID)
    t.Log("Alarm name:    " + globalConfig.AlarmName)

    return nil
}

func enableAlarm(sess *session.Session, instanceName *string, instanceID *string, alarmName *string) error {
    // Create new CloudWatch client
    // snippet-start:[cloudwatch.go.enable_alarm.put]
    svc := cloudwatch.New(sess)

    _, err := svc.PutMetricAlarm(&cloudwatch.PutMetricAlarmInput{
        AlarmName:          alarmName,
        ComparisonOperator: aws.String(cloudwatch.ComparisonOperatorGreaterThanThreshold),
        EvaluationPeriods:  aws.Int64(1),
        MetricName:         aws.String("CPUUtilization"),
        Namespace:          aws.String("AWS/EC2"),
        Period:             aws.Int64(60),
        Statistic:          aws.String(cloudwatch.StatisticAverage),
        Threshold:          aws.Float64(70.0),
        ActionsEnabled:     aws.Bool(true),
        AlarmDescription:   aws.String("Alarm when server CPU exceeds 70%"),
        Unit:               aws.String(cloudwatch.StandardUnitSeconds),
        AlarmActions: []*string{
            aws.String(fmt.Sprintf("arn:aws:swf:"+*svc.Config.Region+":%s:action/actions/AWS_EC2.InstanceId.Reboot/1.0", *instanceName)),
        },
        Dimensions: []*cloudwatch.Dimension{
            {
                Name:  aws.String("InstanceId"),
                Value: instanceID,
            },
        },
    })
    // snippet-end:[cloudwatch.go.enable_alarm.put]
    if err != nil {
        return err
    }

    // Enable the alarm for the instance
    // snippet-start:[cloudwatch.go.enable_alarm.enable]
    _, err = svc.EnableAlarmActions(&cloudwatch.EnableAlarmActionsInput{
        AlarmNames: []*string{
            instanceID,
        },
    })
    // snippet-end:[cloudwatch.go.enable_alarm.enable]
    if err != nil {
        return err
    }

    return nil
}

func disableAlarm(sess *session.Session, alarmName *string) error {
    // Create new CloudWatch client.
    // snippet-start:[cloudwatch.go.disable_call]
    svc := cloudwatch.New(sess)

    _, err := svc.DisableAlarmActions(&cloudwatch.DisableAlarmActionsInput{
        AlarmNames: []*string{
            alarmName,
        },
    })
    // snippet-end:[cloudwatch.go.disable_call]
    if err != nil {
        return err
    }

    return nil
}

func TestDisableAlarmActions(t *testing.T) {
    // When the test started
    thisTime := time.Now()
    nowString := thisTime.Format("2006-01-02 15:04:05 Monday")
    t.Log("Started unit test at " + nowString)

    // Initialize a session that the SDK uses to load
    // credentials from the shared credentials file ~/.aws/credentials
    // and configuration from the shared configuration file ~/.aws/config.
    sess := session.Must(session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    }))

    // Get configuration values
    err := populateConfig(sess, t)
    if err != nil {
        msg := "Could not get configuration values from " + configFileName
        t.Fatal(msg)
    }

    // Create an alarm to disable
    err = enableAlarm(sess, &globalConfig.InstanceName, &globalConfig.InstanceID, &globalConfig.AlarmName)
    if err != nil {
        msg := "Could not create alarm " + globalConfig.AlarmName + " for instance " + globalConfig.InstanceName
        t.Fatal(msg)
    }

    t.Log("Enabled alarm " + globalConfig.AlarmName + " for instance " + globalConfig.InstanceName)

    // Disable alarm
    err = disableAlarm(sess, &globalConfig.AlarmName)
    if err != nil {
        t.Fatal(err)
    }

    t.Log("Disabled alarm " + globalConfig.AlarmName)

    // Finally delete it
    err = DeleteAlarm(sess, &globalConfig.AlarmName)
    if err != nil {
        t.Fatal(err)
    }

    t.Log("Deleted alarm " + globalConfig.AlarmName)
}
