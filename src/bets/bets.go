package bets

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type CreateBetResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type BetObj struct {
	Ticket     string `json:"ticket"`
	TimeStamp  string `json:"timestamp"`
	BetType    string `json:"betType"`
	HomeUserId string `json:"homeUserId"`
	AwayUserId string `json:"awayUserId"`
	History    string `json:"history"`
}

func GetAllBets() (retBets []BetObj, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		return retBets, err
	}

	svc := dynamodb.New(sess)
	proj := expression.NamesList(expression.Name("ticket"), expression.Name("betType"), expression.Name("timestamp"), expression.Name("homeUserId"), expression.Name("awayUserId"), expression.Name("history"))

	expr, err := expression.NewBuilder().WithProjection(proj).Build()

	if err != nil {
		return retBets, err
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String("bets"),
	}

	result, err := svc.Scan(params)
	if err != nil {
		return retBets, err
	}

	for _, i := range result.Items {
		bet := BetObj{}
		err = dynamodbattribute.UnmarshalMap(i, &bet)

		if err != nil {
			return retBets, err
		}
		retBets = append(retBets, bet)
	}
	return retBets, nil
}

func CreateBet(ticket string, timeStamp string, betType string, homeUserId string, awayUserId string, history string) (response CreateBetResponse, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	if err != nil {
		return response, err
	}
	svc := dynamodb.New(sess)
	ticket = strings.Replace(ticket, "_home", "", 1)
	ticket = strings.Replace(ticket, "_away", "", 1)
	bet := &BetObj{Ticket: ticket, TimeStamp: timeStamp, BetType: betType, HomeUserId: homeUserId, AwayUserId: awayUserId, History: history}
	atts, err := dynamodbattribute.MarshalMap(bet)

	if err != nil {
		return response, err
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{Item: atts, TableName: aws.String("bets")})

	if err != nil {
		return response, err
	}
	response.Status = true
	response.Message = "Bet created: " + ticket + "_" + timeStamp
	return response, nil
}
