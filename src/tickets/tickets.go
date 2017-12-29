package tickets

import (
	"apblogger"
	"encoding/json"
	"user"
	// "bets"
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

// if the ticket is created, try to make a bet
func createBetFromTicketInNewRoutine(ticket string, key string) {
	apblogger.LogMessage("createBetFromTicketInNewRoutine Enter ticket=" + ticket + " key=" + key)
	time.Sleep(5 * time.Second)
	apblogger.LogMessage("createBetFromTicketInNewRoutine Post Sleep")
	CreateBetFromTicketID(ticket, key)
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
		go createBetFromTicketInNewRoutine(ticketString, sortKey)
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

type BetResponse struct {
	Ticket  TicketObj `json:"ticket"`
	Oppose  TicketObj `json:"opposet"`
	Status  bool      `json:"status"`
	Message string    `json:"message"`
}

func CreateBetFromTicketID(ticketStr string, sortKey string) (wasCreated bool, err error) {
	apblogger.LogMessage("CreateBetFromTicketID")
	ticket, err := GetTicket(ticketStr, sortKey)
	if err != nil {
		apblogger.LogVar("err", fmt.Sprintf("%v", err))
		return false, err
	} else {
		if ticket.Ticket == "" {
			return false, errors.New("Ticket can't be found")
		} else {
			apblogger.LogMessage("Found Ticket")
			ticketsOppose, err := GetTicketOppose(ticket.Ticket)
			if err != nil {
				return false, err
			}
			if len(ticketsOppose) == 0 {
				apblogger.LogMessage("No Opposing Ticket")
				return false, errors.New("No Opposing Ticket")
			} else {
				apblogger.LogMessage("Found potential opposition")
				ticketOpposeTop := ticketsOppose[0]
				apblogger.LogVar("ticketOpposeTop", fmt.Sprintf("%v", ticketOpposeTop))
				ticketPrice, err := strconv.ParseInt(ticket.Price, 10, 0)
				apblogger.LogVar("ticketPrice", fmt.Sprintf("%v", ticketPrice))
				ticketPriceOppose, err := strconv.ParseInt(ticketOpposeTop.Price, 10, 0)
				apblogger.LogVar("ticketPriceOppose", fmt.Sprintf("%v", ticketPriceOppose))
				maxPrice, err := strconv.ParseInt("100", 10, 0)
				if err != nil {
					apblogger.LogVar("err", fmt.Sprintf("%v", err))
					return false, err
				}

				if maxPrice-ticketPrice-ticketPriceOppose > 0 {
					apblogger.LogMessage("The prices didn't equal")
					return false, errors.New("The prices didn't equal")
				} else {
					apblogger.LogMessage("Looking good")
					overPaid, err := strconv.ParseInt("0", 10, 0)
					if maxPrice != ticketPrice+ticketPriceOppose {
						overPaid = maxPrice - ticketPrice - ticketPriceOppose
					}
					ticketPriceAdjusted := int64(ticketPrice + (overPaid / 2))
					ticketPriceOpposeAdjusted := maxPrice - ticketPriceAdjusted
					ticketToWrite := ticket
					ticketToWriteOpposeTop := ticketOpposeTop

					ticketToWrite.PriceAdjusted = fmt.Sprintf("%v", ticketPriceAdjusted)
					ticketToWriteOpposeTop.PriceAdjusted = fmt.Sprintf("%v", ticketPriceOpposeAdjusted)

					userTicket := user.GetUser(ticket.UserID)
					userTicketBankroll := int64(userTicket.Bankroll)
					userTicketOppose := user.GetUser(ticketOpposeTop.UserID)
					userTicketOpposeBankroll := int64(userTicketOppose.Bankroll)
					if userTicketBankroll < ticketPriceAdjusted {
						return false, errors.New("User doesn't have enough bankroll")
					}
					if userTicketOpposeBankroll < ticketPriceOpposeAdjusted {
						return false, errors.New("User 2 doesn't have enough bankroll")
					}

					history := BetResponse{Ticket: ticketToWrite, Oppose: ticketToWriteOpposeTop}
					historyJSON, _ := json.Marshal(history)
					timestamp := fmt.Sprintf("%v", time.Now().UnixNano())

					apblogger.LogMessage("Abount to create Bet " + ticket.Ticket + ": " + ticket.UserID + " v.s.v" + ticketOpposeTop.UserID)
					err = CreateBet(ticket.Ticket, timestamp, ticket.BetType, ticket.UserID, ticketOpposeTop.UserID, string(historyJSON))
					if err != nil {
						return false, err
					} else {

						user.UpdateUser(ticket.UserID, fmt.Sprintf("%v", userTicketBankroll-ticketPriceAdjusted))
						user.UpdateUser(ticketOpposeTop.UserID, fmt.Sprintf("%v", userTicketOpposeBankroll-ticketPriceOpposeAdjusted))

						apblogger.LogMessage("Abount to delete ticket1")
						err = DeleteTicket(ticket.Ticket, ticket.PriceAndTime)
						apblogger.LogVar("del ticket 1 err", fmt.Sprintf("%v", err))

						apblogger.LogMessage("Abount to delete ticket2")
						err = DeleteTicket(ticketOpposeTop.Ticket, ticketOpposeTop.PriceAndTime)
						apblogger.LogVar("del ticket 2 err", fmt.Sprintf("%v", err))

						return true, nil
					}
				}

			}
		}
	}
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

func CreateBet(ticket string, timeStamp string, betType string, homeUserId string, awayUserId string, history string) (err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	if err != nil {
		return err
	}
	svc := dynamodb.New(sess)
	ticket = utils.GetTicketWithoutSide(ticket)
	bet := &BetObj{Ticket: ticket, TimeStamp: timeStamp, BetType: betType, HomeUserId: homeUserId, AwayUserId: awayUserId, History: history}

	atts, err := dynamodbattribute.MarshalMap(bet)

	if err != nil {
		return err
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{Item: atts, TableName: aws.String("bets")})

	if err != nil {
		return err
	} else {
		stats.UpdateCounter(ticket, "bets", "1")
		return nil
	}
}
