package bets

import (
	"encoding/json"
	"errors"
	"fmt"
	"stats"
	"strconv"
	"tickets"
	"time"
	"user"
	"utils"

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

type BetResponse struct {
	Ticket  tickets.TicketObj `json:"ticket"`
	Oppose  tickets.TicketObj `json:"opposet"`
	Status  bool              `json:"status"`
	Message string            `json:"message"`
}

func CreateBetFromTicketID(ticketStr string, sortKey string) (wasCreated bool, err error) {
	ticket, error := tickets.GetTicket(ticketStr, sortKey)
	if err != nil {
		return false, error
	} else {
		if ticket.Ticket == "" {
			return false, errors.New("Ticket can't be found")
		} else {
			ticketsOppose, err := tickets.GetTicketOppose(ticket.Ticket)
			if err != nil {
				return false, err
			}
			if len(ticketsOppose) == 0 {
				return false, errors.New("No Opposing Ticket")
			} else {
				ticketOpposeTop := ticketsOppose[0]
				ticketPrice, err := strconv.ParseInt(ticket.Price, 10, 0)
				ticketPriceOppose, err := strconv.ParseInt(ticketOpposeTop.Price, 10, 0)
				maxPrice, err := strconv.ParseInt("100", 10, 0)
				if err != nil {
					return false, err
				}

				if maxPrice-ticketPrice-ticketPriceOppose > 0 {
					return false, errors.New("The prices didn't equal")
				} else {
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

					err = CreateBet(ticket.Ticket, timestamp, ticket.BetType, ticket.UserID, ticketOpposeTop.UserID, string(historyJSON))
					if err != nil {
						return false, err
					} else {

						user.UpdateUser(ticket.UserID, fmt.Sprintf("%v", userTicketBankroll-ticketPriceAdjusted))
						user.UpdateUser(ticketOpposeTop.UserID, fmt.Sprintf("%v", userTicketOpposeBankroll-ticketPriceOpposeAdjusted))
						err = tickets.DeleteTicket(ticket.Ticket, ticket.PriceAndTime)
						err = tickets.DeleteTicket(ticketOpposeTop.Ticket, ticketOpposeTop.PriceAndTime)
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
