DynamoDB Demo Project
-----

Run LocalStack:
```
docker-compose -f docker-compose-dynamodb.yaml up -d
```

List tables:
```
aws --endpoint-url=http://localhost:4566 --region=eu-central-1 dynamodb list-tables
```