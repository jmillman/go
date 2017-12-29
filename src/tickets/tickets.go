package tickets

import (
	"errors"
	"fmt"
	"stats"
	"strconv"
	"strings"
	"time"
	"utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// CreateTicketResponse response for the create ticket api
type CreateTicketResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// TicketObj object for the ticket
type TicketObj struct {
	Ticket        string `json:"ticket"`
	BetType       string `json:"betType"`
	Side          string `json:"side"`
	Price         string `json:"price"`
	PriceAdjusted string `json:"priceAdjusted"`
	UserID        string `json:"userId"`
	PriceAndTime  string `json:"priceAndTime"`
}

// DeleteTicket deletes ticket from DB
func DeleteTicket(ticket string, priceAndTime string) (err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("tickets"),
		Key: map[string]*dynamodb.AttributeValue{
			"ticket": {
				S: aws.String(ticket),
			},
			"priceAndTime": {
				S: aws.String(priceAndTime),
			},
		},
	}

	_, err = svc.DeleteItem(input)

	return err

	// err = dynamodbattribute.UnmarshalMap(result.Item, &retTicket)
	//
	// if err != nil {
	// 	return retTicket, err
	// 	// panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	// }
	// return retTicket, nil
}

// GetTicket gets the ticket from the db
func GetTicket(ticket string, priceAndTime string) (retTicket TicketObj, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		return retTicket, err
	}

	svc := dynamodb.New(sess)
	input := &dynamodb.GetItemInput{
		TableName: aws.String("tickets"),
		Key: map[string]*dynamodb.AttributeValue{
			"ticket": {
				S: aws.String(ticket),
			},
			"priceAndTime": {
				S: aws.String(priceAndTime),
			},
		},
	}

	result, err := svc.GetItem(input)

	if err != nil {
		return retTicket, err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &retTicket)

	if err != nil {
		return retTicket, err
		// panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	return retTicket, nil
}

// GetTicketOppose gets the away for a call and the call for an away
func GetTicketOppose(ticket string) (retTickets []TicketObj, err error) {
	hasSide := false
	if strings.Contains(ticket, "_home") {
		ticket = strings.Replace(ticket, "_home", "_away", 1)
		hasSide = true
	} else if strings.Contains(ticket, "_away") {
		ticket = strings.Replace(ticket, "_away", "_home", 1)
		hasSide = true
	}
	if !hasSide {
		return retTickets, errors.New("Ticket has no side")
	}

	return GetTicketCommon(ticket)
}

// GetTicketCommon gets all the same tickets, meaning all the same home or aways
func GetTicketCommon(ticket string) (retTickets []TicketObj, err error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	if err != nil {
		return retTickets, err
	}

	svc := dynamodb.New(sess)

	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String("tickets"),
		KeyConditions: map[string]*dynamodb.Condition{
			"ticket": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(ticket),
					},
				},
			},
		},
	}

	result, err := svc.Query(queryInput)

	if err != nil {
		return retTickets, err
	}

	for _, i := range result.Items {
		ticket := TicketObj{}
		err = dynamodbattribute.UnmarshalMap(i, &ticket)
		if err != nil {
			return retTickets, err
		}
		retTickets = append(retTickets, ticket)
	}
	return retTickets, nil
}

// GetAllTickets gets all tickets from the database
func GetAllTickets() (retTickets []TicketObj, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	fmt.Printf("%v\n", err)

	if err != nil {
		return retTickets, err
	}

	fmt.Printf("%v\n", err)

	svc := dynamodb.New(sess)
	proj := expression.NamesList(expression.Name("ticket"), expression.Name("betType"), expression.Name("side"), expression.Name("price"), expression.Name("userId"), expression.Name("priceAndTime"))

	expr, err := expression.NewBuilder().WithProjection(proj).Build()

	fmt.Printf("%v\n", err)

	if err != nil {
		return retTickets, err
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String("tickets"),
	}

	result, err := svc.Scan(params)
	if err != nil {
		return retTickets, err
	}

	for _, i := range result.Items {
		ticket := TicketObj{}
		err = dynamodbattribute.UnmarshalMap(i, &ticket)

		if err != nil {
			return retTickets, err
		}
		retTickets = append(retTickets, ticket)
	}
	return retTickets, nil
}

// CreateTicketsFromFormData creates a ticket
func CreateTicketsFromFormData(ticketString string, betType string, side string, priceStr string, userID string, quantityStr string) (response CreateTicketResponse, err error) {
	quantity, _ := strconv.ParseInt(quantityStr, 10, 0)
	// TODO: don't hardcode the max Value
	price, _ := strconv.ParseInt(priceStr, 10, 0)
	sortKeyPrice := 100 - price //need to sort by largest number and dynamodb sorts by smallest so have to take invese, should be max - price, hard coded in 100
	for j := int(1); j <= int(quantity); j++ {
		sortKey := fmt.Sprintf("%v_%v_%vOF%v", sortKeyPrice, time.Now().UnixNano(), j, quantity)
		_, err = createTicket(sortKey, ticketString, betType, side, priceStr, userID)
		if err != nil {
			return response, err
		}
	}
	return response, err
}

// createTicket creates a ticket
func createTicket(priceAndTime string, ticketString string, betType string, side string, price string, userID string) (response CreateTicketResponse, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	if err != nil {
		return response, err
	}
	svc := dynamodb.New(sess)
	ticket := &TicketObj{PriceAndTime: priceAndTime, Ticket: ticketString, BetType: betType, Side: side, Price: price, UserID: userID}
	atts, err := dynamodbattribute.MarshalMap(ticket)

	if err != nil {
		return response, err
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{Item: atts, TableName: aws.String("tickets")})

	if err != nil {
		return response, err
	} else {
		// ticket created
		ticketWithoutSide := utils.GetTicketWithoutSide(ticketString)

		stats.UpdateStatIfGreater(ticketWithoutSide, side, price)

		response.Status = true
		response.Message = ticketString
		return response, nil
	}
}
