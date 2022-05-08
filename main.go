package main

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DemoApp struct {
    ctx context.Context
    svc *dynamodb.Client
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
        ctx: context.TODO(),
    }
    app.initClient()

    tableName := "LikedSongs"
    app.deleteTable(tableName)
    app.listTables()

    app.createTable()
    app.listTables()
}

func (app *DemoApp) listTables() {
    resp, err := app.svc.ListTables(app.ctx, &dynamodb.ListTablesInput{
        Limit: aws.Int32(5),
    })
    if err != nil {
        log.Fatalf("failed to list tables, %v", err)
    }

    fmt.Println("Tables: [")
    for _, tableName := range resp.TableNames {
        fmt.Println(tableName)
    }
    fmt.Println("]")
}

func (app *DemoApp) createTable() {
    tableName := "LikedSongs"

    input := &dynamodb.CreateTableInput{
        AttributeDefinitions: []types.AttributeDefinition{
            {
                AttributeName: aws.String("Artist"),
                AttributeType: types.ScalarAttributeTypeS,
            },
            {
                AttributeName: aws.String("Title"),
                AttributeType: types.ScalarAttributeTypeS,
            },
        },
        KeySchema: []types.KeySchemaElement{
            {
                AttributeName: aws.String("Artist"),
                KeyType:       types.KeyTypeHash,
            },
            {
                AttributeName: aws.String("Title"),
                KeyType:       types.KeyTypeRange,
            },
        },
        ProvisionedThroughput: &types.ProvisionedThroughput{
            ReadCapacityUnits:  aws.Int64(5),
            WriteCapacityUnits: aws.Int64(5),
        },
        TableName: aws.String(tableName),
    }

    _, err := app.svc.CreateTable(app.ctx, input)
    if err != nil {
        log.Fatalf("Got error calling CreateTable: %s", err)
    }

    fmt.Println("Created the table:", tableName)
}

func (app *DemoApp) deleteTable(tableName string) {
    _, err := app.svc.DeleteTable(app.ctx, &dynamodb.DeleteTableInput{
        TableName: aws.String(tableName),
    })
    if err != nil {
        log.Fatalf("Got error calling DeleteTable: %s", err)
    }

    fmt.Println("Deleted the table:", tableName)
}