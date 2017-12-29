package stats

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

//StatObj represets stats table
type StatObj struct {
	Ticket    string `json:"ticket"`
	StatType  string `json:"statType"`
	StatValue string `json:"statValue"`
}

var statsTableName = "stats2"

//GetAllStats returns all stats from the database
func GetAllStats() (retStats []StatObj) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := dynamodb.New(sess)
	proj := expression.NamesList(expression.Name("ticket"), expression.Name("statType"), expression.Name("statValue"))
	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String(statsTableName),
	}

	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, i := range result.Items {
		stat := StatObj{}
		err = dynamodbattribute.UnmarshalMap(i, &stat)
		retStats = append(retStats, stat)
	}
	return retStats
}

// func GetStat(ticket string) (retStat StatObj) {
// 	sess, err := session.NewSession(&aws.Config{
// 		Region: aws.String("us-east-2")},
// 	)
//
// 	svc := dynamodb.New(sess)
//
// 	result, err := svc.GetItem(&dynamodb.GetItemInput{
// 		TableName: aws.String(statsTableName),
// 		Key: map[string]*dynamodb.AttributeValue{
// 			"ticket": {
// 				S: aws.String(ticket),
// 			},
// 		},
// 	})
//
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
//
// 	err = dynamodbattribute.UnmarshalMap(result.Item, &retStat)
//
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
// 	}
// 	return retStat
// }

func getStatValue(ticket string, statType string) (statValue string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(statsTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ticket": {
				S: aws.String(ticket),
			},
			"statType": {
				S: aws.String(statType),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	var retStat StatObj
	err = dynamodbattribute.UnmarshalMap(result.Item, &retStat)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	return retStat.StatValue
}

// UpdateStatResponse object to be returned from the update stat call
type UpdateStatResponse struct {
	Error   bool   `json:"error"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// UpdateStatIfGreater updates a stat if greater than the old
func UpdateStatIfGreater(ticket string, statType string, statValue string) (updateStatResponse UpdateStatResponse, err error) {
	currentStatValue := getStatValue(ticket, statType)
	oldVal, _ := strconv.ParseInt(currentStatValue, 10, 0)
	newVal, _ := strconv.ParseInt(statValue, 10, 0)
	if newVal > oldVal {
		return UpdateStat(ticket, statType, statValue)
	}
	return UpdateStatResponse{Error: false, Message: "Stat not greater and not updated"}, nil
}

// UpdateStat updates a stat in the database
func UpdateStat(ticket string, statType string, statValue string) (updateStatResponse UpdateStatResponse, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := dynamodb.New(sess)

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#V": aws.String("statValue"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				N: aws.String(statValue),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"ticket": {
				S: aws.String(ticket),
			},
			"statType": {
				S: aws.String(statType),
			},
		},
		// leave this out because we want it created if it doesn't exist
		// ConditionExpression: aws.String("attribute_exists()"), //this is because if it has email it exists, if not it would be a new record and dynamo will just enter it
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        aws.String(statsTableName),
		UpdateExpression: aws.String("SET #V = :v"),
	}

	_, err = svc.UpdateItem(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				updateStatResponse.Error = true
				updateStatResponse.Message = "Item does not exist"
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
				return updateStatResponse, err
			default:
				fmt.Println(aerr.Error())
				return updateStatResponse, aerr
			}
		} else {
			fmt.Println(err.Error())
		}

		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return updateStatResponse, err
}

// UpdateCounter updates a stat in the database
func UpdateCounter(ticket string, statType string, amount string) (updateStatResponse UpdateStatResponse, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := dynamodb.New(sess)

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":num": {
				N: aws.String(amount),
			},
			":start": {
				N: aws.String("0"),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"ticket": {
				S: aws.String(ticket),
			},
			"statType": {
				S: aws.String(statType),
			},
		},
		// leave this out because we want it created if it doesn't exist
		// ConditionExpression: aws.String("attribute_exists()"), //this is because if it has email it exists, if not it would be a new record and dynamo will just enter it
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        aws.String(statsTableName),
		UpdateExpression: aws.String("SET statValue = if_not_exists(statValue, :start) + :num"), //increment counter by val passed in, if no counter, set to val passed in
		// UpdateExpression: aws.String("SET statValue = statValue + :num"),
	}
	_, err = svc.UpdateItem(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				updateStatResponse.Error = true
				updateStatResponse.Message = "Item does not exist"
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
				return updateStatResponse, err
			default:
				fmt.Println(aerr.Error())
				return updateStatResponse, aerr
			}
		} else {
			fmt.Println(err.Error())
		}

		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return updateStatResponse, err
}
