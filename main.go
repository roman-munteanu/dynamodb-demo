package main

import (
    "context"
    "fmt"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DemoApp struct {
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

    cfg, err := config.LoadDefaultConfig(context.TODO(),
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

    app := DemoApp{}
    app.initClient()
    
    resp, err := app.svc.ListTables(context.TODO(), &dynamodb.ListTablesInput{
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
