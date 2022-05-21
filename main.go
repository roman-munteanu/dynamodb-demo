package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DemoApp struct {
	ctx       context.Context
	svc       *dynamodb.Client
	tableName string
}

type Song struct {
	Artist      string
	ReleaseDate string
	Title       string
	Genre       string
}

func (app *DemoApp) initClient() {
	theRegion := "eu-central-1"

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == dynamodb.ServiceID && region == theRegion {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://localhost:4566",
				SigningRegion: theRegion,
			}, nil
		}

		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(app.ctx,
		config.WithRegion(theRegion),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("localstack", "localstack", "session")),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	app.svc = dynamodb.NewFromConfig(cfg)
}

func main() {

	app := DemoApp{
		ctx:       context.TODO(),
		tableName: "LikedSongs",
	}
	app.initClient()

	app.deleteTable()
	app.listTables()

	app.createTable()
	app.listTables()

	app.putItem(Song{"RMHighlander", "2021-11-13", "Odyssey", "Indie"})
	app.putItem(Song{"RMHighlander", "2022-04-28", "Pure Shore", "Travel"})
	app.putItem(Song{"RMHighlander", "2021-10-30", "Steady Flight", "Electric Blues"})

	// app.scan()

	// app.getItem("RMHighlander", "2021-11-13")

	// app.updateItem("RMHighlander", "2021-11-13", "Soft Rock")

	// app.getItem("RMHighlander", "2021-11-13")

	// app.query("RMHighlander", "2022-01-01")

	// app.deleteItem("RMHighlander", "2021-10-30")

	// app.scan()

	app.transactGetItems()

	app.transactWriteItems()

	app.scan()
}

func (app *DemoApp) listTables() {
	fmt.Println("--------- listTables:")
	resp, err := app.svc.ListTables(app.ctx, &dynamodb.ListTablesInput{
		Limit: aws.Int32(5),
	})
	if err != nil {
		log.Fatalf("failed to list tables, %v", err)
	}

	fmt.Println("Tables:")
	for _, tableName := range resp.TableNames {
		fmt.Println(tableName)
	}
}

func (app *DemoApp) createTable() {
	fmt.Println("--------- createTable:")

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("Artist"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("ReleaseDate"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("Artist"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("ReleaseDate"),
				KeyType:       types.KeyTypeRange,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(app.tableName),
	}

	_, err := app.svc.CreateTable(app.ctx, input)
	if err != nil {
		log.Fatalf("Got error calling CreateTable: %s", err)
		return
	}

	fmt.Println("Created the table:", app.tableName)
}

func (app *DemoApp) deleteTable() {
	fmt.Println("--------- deleteTable:")
	_, err := app.svc.DeleteTable(app.ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(app.tableName),
	})
	if err != nil {
		log.Fatalf("Got error calling DeleteTable: %s", err)
		return
	}

	fmt.Println("Deleted the table:", app.tableName)
}

func (app *DemoApp) putItem(song Song) {
	fmt.Println("--------- putItem:")

	// alternative: item, err := attributevalue.MarshalMap(song)
	_, err := app.svc.PutItem(app.ctx, &dynamodb.PutItemInput{
		TableName: aws.String(app.tableName),
		Item: map[string]types.AttributeValue{
			"Artist":      &types.AttributeValueMemberS{Value: song.Artist},
			"ReleaseDate": &types.AttributeValueMemberS{Value: song.ReleaseDate},
			"Title":       &types.AttributeValueMemberS{Value: song.Title},
			"Genre":       &types.AttributeValueMemberS{Value: song.Genre},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
		return
	}

	fmt.Println("PutItem:", song.Artist, song.ReleaseDate)
}

func (app *DemoApp) updateItem(artist, releaseDate, genre string) {
	fmt.Println("--------- updateItem:")
	var attrs map[string]string

	resp, err := app.svc.UpdateItem(app.ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(app.tableName),
		Key: map[string]types.AttributeValue{
			"Artist":      &types.AttributeValueMemberS{Value: artist},
			"ReleaseDate": &types.AttributeValueMemberS{Value: releaseDate},
		},
		UpdateExpression: aws.String("SET Genre = :val"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":val": &types.AttributeValueMemberS{Value: genre},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling UpdateItem: %s", err)
		return
	}

	err = attributevalue.UnmarshalMap(resp.Attributes, &attrs)
	if err != nil {
		log.Fatalf("Unmarshal error: %s", err)
	}

	fmt.Println(attrs)
}

func (app *DemoApp) scan() {
	fmt.Println("--------- scan:")
	var items []Song

	resp, err := app.svc.Scan(app.ctx, &dynamodb.ScanInput{
		TableName: aws.String(app.tableName),
	})
	if err != nil {
		log.Fatalf("Got error calling Scan: %s", err)
		return
	}

	err = attributevalue.UnmarshalListOfMaps(resp.Items, &items)
	if err != nil {
		log.Fatalf("Unmarshal error: %s", err)
	}

	fmt.Println(items)
}

func (app *DemoApp) getItem(artist, releaseDate string) {
	fmt.Println("--------- getItem:")
	var item Song

	resp, err := app.svc.GetItem(app.ctx, &dynamodb.GetItemInput{
		TableName: aws.String(app.tableName),
		Key: map[string]types.AttributeValue{
			"Artist":      &types.AttributeValueMemberS{Value: artist},
			"ReleaseDate": &types.AttributeValueMemberS{Value: releaseDate},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling GetItem: %s", err)
		return
	}

	err = attributevalue.UnmarshalMap(resp.Item, &item)
	if err != nil {
		log.Fatalf("Unmarshal error: %s", err)
		return
	}

	fmt.Println(item)
}

func (app *DemoApp) query(artist, date string) {
	fmt.Println("--------- query:")
	var items []Song

	resp, err := app.svc.Query(app.ctx, &dynamodb.QueryInput{
		TableName:              aws.String(app.tableName),
		KeyConditionExpression: aws.String("Artist = :a AND #date > :d"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":a": &types.AttributeValueMemberS{Value: artist},
			":d": &types.AttributeValueMemberS{Value: date},
		},
		ExpressionAttributeNames: map[string]string{
			"#date": "ReleaseDate",
		},
	})
	if err != nil {
		log.Fatalf("Got error calling Query: %s", err)
		return
	}

	err = attributevalue.UnmarshalListOfMaps(resp.Items, &items)
	if err != nil {
		log.Fatalf("Unmarshal error: %s", err)
	}

	fmt.Println(items)
}

func (app *DemoApp) deleteItem(artist, releaseDate string) {
	fmt.Println("--------- deleteItem:")

	_, err := app.svc.DeleteItem(app.ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(app.tableName),
		Key: map[string]types.AttributeValue{
			"Artist":      &types.AttributeValueMemberS{Value: artist},
			"ReleaseDate": &types.AttributeValueMemberS{Value: releaseDate},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling DeleteItem: %s", err)
		return
	}

	fmt.Println("DeleteItem: ", artist, releaseDate)
}

func (app *DemoApp) transactGetItems() {
	fmt.Println("--------- transactGetItems:")

	resp, err := app.svc.TransactGetItems(app.ctx, &dynamodb.TransactGetItemsInput{
		TransactItems: []types.TransactGetItem{
			{
				Get: &types.Get{
					TableName: aws.String(app.tableName),
					Key: map[string]types.AttributeValue{
						"Artist":      &types.AttributeValueMemberS{Value: "RMHighlander"},
						"ReleaseDate": &types.AttributeValueMemberS{Value: "2021-11-13"},
					},
				},
			},
			{
				Get: &types.Get{
					TableName: aws.String(app.tableName),
					Key: map[string]types.AttributeValue{
						"Artist":      &types.AttributeValueMemberS{Value: "RMHighlander"},
						"ReleaseDate": &types.AttributeValueMemberS{Value: "2021-10-30"},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling Query: %s", err)
		return
	}

	for _, itemResponse := range resp.Responses {
		var item Song

		err = attributevalue.UnmarshalMap(itemResponse.Item, &item)
		if err != nil {
			log.Fatalf("Unmarshal error: %s", err)
			break
		}

		fmt.Println(item)
	}
}

func (app *DemoApp) transactWriteItems() {
	fmt.Println("--------- transactWriteItems:")

	resp, err := app.svc.TransactWriteItems(app.ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName: aws.String(app.tableName),
					Key: map[string]types.AttributeValue{
						"Artist":      &types.AttributeValueMemberS{Value: "RMHighlander"},
						"ReleaseDate": &types.AttributeValueMemberS{Value: "2022-04-28"},
					},
				},
			},
			{
				Update: &types.Update{
					TableName: aws.String(app.tableName),
					Key: map[string]types.AttributeValue{
						"Artist":      &types.AttributeValueMemberS{Value: "RMHighlander"},
						"ReleaseDate": &types.AttributeValueMemberS{Value: "2021-11-13"},
					},
					UpdateExpression: aws.String("SET Genre = :newval"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":newval": &types.AttributeValueMemberS{Value: "Rock"},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling Query: %s", err)
		return
	}

	fmt.Println(resp)
}
