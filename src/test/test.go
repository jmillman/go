/*
   Copyright 2010-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
    "fmt"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type User struct {
  Id string`json:"Id"`
  Email string`json:"email"`
  Name string`json:"name"`
}

func main() {
    // Initialize a session in us-west-2 that the SDK will use to load
    // credentials from the shared credentials file ~/.aws/credentials.
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-2")},
    )

    // Create DynamoDB client
    svc := dynamodb.New(sess)

    result, err := svc.GetItem(&dynamodb.GetItemInput{
        TableName: aws.String("Users"),
        Key: map[string]*dynamodb.AttributeValue{
            "Id": {
                S: aws.String("2"),
            },
        },
    })

    if err != nil {
        fmt.Println(err.Error())
        return
    }

    user := User{}

    err = dynamodbattribute.UnmarshalMap(result.Item, &user)

    if err != nil {
        panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
    }

    if user.Name == "" {
        fmt.Println("Could not find user")
        return
    }

    fmt.Println("Found item:")
    fmt.Println("name:  ", user.Name)
    fmt.Println("email:  ", user.Email)
    fmt.Println("Id:  ", user.Id)
}
