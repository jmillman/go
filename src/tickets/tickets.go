package tickets

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type CreateTicketResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type TicketObj struct {
	Ticket       string `json:"ticket"`
	BetType      string `json:"betType"`
	Side         string `json:"side"`
	Price        string `json:"price"`
	UserId       string `json:"userId"`
	PriceAndTime string `json:"priceAndTime"`
}

func GetAllTickets() (retTickets []TicketObj) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	svc := dynamodb.New(sess)
	proj := expression.NamesList(expression.Name("ticket"), expression.Name("betType"), expression.Name("side"), expression.Name("price"), expression.Name("userId"), expression.Name("priceAndTime"))

	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String("tickets"),
	}

	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, i := range result.Items {
		ticket := TicketObj{}
		err = dynamodbattribute.UnmarshalMap(i, &ticket)
		retTickets = append(retTickets, ticket)
	}
	return retTickets
}

func CreateTicket(priceAndTime string, ticketString string, betType string, side string, price string, userId string) (response CreateTicketResponse) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	svc := dynamodb.New(sess)
	ticket := &TicketObj{PriceAndTime: priceAndTime, Ticket: ticketString, BetType: betType, Side: side, Price: price, UserId: userId}
	atts, err := dynamodbattribute.MarshalMap(ticket)

	if err != nil {
		log.Panic(err)
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{Item: atts, TableName: aws.String("tickets")})

	response.Status = true
	response.Message = "Ticket Created"
	return response
}
