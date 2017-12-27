package user

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type UserObj struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Bankroll int    `json:"bankroll"`
}

type ListOfUsers struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	Status   bool   `json:"status"`
	Message  string `json:"message"`
	Email    string `json:"userId"`
	Bankroll int    `json:"bankroll"`
}

func GetAllUsers() (retUsers []UserObj) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := dynamodb.New(sess)
	proj := expression.NamesList(expression.Name("email"), expression.Name("password"))
	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String("Users"),
	}

	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, i := range result.Items {
		user := UserObj{}
		err = dynamodbattribute.UnmarshalMap(i, &user)
		retUsers = append(retUsers, user)
	}
	return retUsers
}

func GetUser(email string) (retUser UserObj) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &retUser)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	return retUser
}

func LoginUser(email string, password string) (response UserResponse) {

	tmpUser := GetUser(email)

	if tmpUser.Password != password {
		response.Status = false
		response.Message = "User Not Found"
	} else {
		response.Status = true
		response.Email = tmpUser.Email
		response.Bankroll = tmpUser.Bankroll
	}
	return response
}

func PutUser(email string, password string) (response UserResponse) {
	tmpUser := GetUser(email)
	// the user doesn't exist yet, so create
	if tmpUser.Email == "" {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-2")},
		)

		svc := dynamodb.New(sess)
		bankRoll := 1000
		user := &UserObj{Email: email, Password: password, Bankroll: bankRoll}
		atts, err := dynamodbattribute.MarshalMap(user)

		if err != nil {
			log.Panic(err)
		}
		_, err = svc.PutItem(&dynamodb.PutItemInput{Item: atts, TableName: aws.String("Users")})

		response.Status = err == nil
		response.Email = email
		response.Bankroll = bankRoll
		response.Message = ""
	} else {
		response.Status = false
		response.Email = email
		response.Message = "User already exists"
	}

	return response
}
