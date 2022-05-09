DynamoDB Demo Project
-----

## Run the app

Run LocalStack:
```
docker-compose -f docker-compose-dynamodb.yaml up -d
```

Build and run:
```
go mod tidy

go run main.go
```


## Brief Overview

### Features:

+ NoSQL
+ Distributed: horizontal scalability
+ Replication across multiple AZs
+ Handles massive workloads
+ Integrated with IAM
+ Streams (supports events)


### Tables

* Tables (table is a collection of data)
* Items (rows, item is a group of attributes) (maximum size: 400KB, no limit to the number of items you can store in a table)
* Attributes (columns, most of them are scalar but could be nested)
* Data types:
	- Scalar types: String, Number, Boolean, Null
	- Document types: Map, List
	- Set types: Number Set, String Set, Binary Set


### Primary keys options:

1. Partition key (HASH) - decided on creation time, unique for each item and diverse
        hash function determines the partition in which the item will be stored
		Example: ArtistID or: SongID 

2. Partition key + Sort key (HASH + RANGE) - composite primary key, must be unique
        All items with the same partition key value are stored together, in sorted order by sort key value.
		Example: SongID (HASH), ReleaseTS (RANGE)
		or: UserID (HASH), PostID (RANGE)


### GSI, LSI
		TODO


### API:

* **GetItem**: read by HASH or HASH+RANGE
		ProjectionExpression: retrieve only certain attributes

* **Query**: retrieve items by a partition key (HASH) or with a sort key (RANGE)
		FilterExpression: by other attributes (by non HASH/RANGE attributes)
		Limit or up to 1 MB
		GSI, LSI

* **Scan**: read the entire table
		up to 1 MB + pagination
		supports parallel scan
		ProjectionExpression, FilterExpression

* **PutItem**: create or replace item

* **UpdateItem**: update the existing item's attributes or create a new item

* **DeleteItem**: delete item 
		or delete by a condition

* **DeleteTable**: delete a table and all items

* **BatchWriteItem**: up to 25 PutItem/DeleteItem (but not UpdateItem)
		up to 16 MB of data

* **BatchGetItem**:
		up to 16 MB


### AWS CLI

* Create a table:
```
aws dynamodb create-table \
    --table-name LikedSongs \
    --attribute-definitions \
        AttributeName=Artist,AttributeType=S \
        AttributeName=ReleaseDate,AttributeType=S \
    --key-schema \
        AttributeName=Artist,KeyType=HASH \
        AttributeName=ReleaseDate,KeyType=RANGE \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url=http://localhost:4566 \
    --region=eu-central-1
```

* List tables:
```
aws dynamodb list-tables \
    --endpoint-url=http://localhost:4566 \
    --region=eu-central-1 

aws dynamodb describe-table \
    --table-name LikedSongs | grep TableStatus \
    --endpoint-url=http://localhost:4566
```

* Drop table:
```
aws dynamodb delete-table \
    --table-name LikedSongs \
    --endpoint-url=http://localhost:4566 
```

* Create data:
```
aws dynamodb put-item \
    --table-name LikedSongs \
    --item \
    '{"Artist": {"S": "RMHighlander"}, "ReleaseDate": {"S": "2021-11-13"}, "Title": {"S": "Odyssey"}, "Genre": {"S": "Indie"}}' \
    --endpoint-url=http://localhost:4566


aws dynamodb put-item \
    --table-name LikedSongs \
    --item \
    '{"Artist": {"S": "RMHighlander"}, "ReleaseDate": {"S": "2022-04-28"}, "Title": {"S": "Pure Shore"}, "Genre": {"S": "Travel"}}' \
    --endpoint-url=http://localhost:4566


aws dynamodb put-item \
    --table-name LikedSongs \
    --item \
    '{"Artist": {"S": "RMHighlander"}, "ReleaseDate": {"S": "2021-10-30"}, "Title": {"S": "Steady Flight"}, "Genre": {"S": "Electric Blues"}}' \
    --endpoint-url=http://localhost:4566
```

* Batch write:
```
aws dynamodb batch-write-item \
    --request-items file://data-write.json \
    --return-consumed-capacity INDEXES \
    --return-item-collection-metrics SIZE \
    --endpoint-url=http://localhost:4566
```

* Update data:
```
aws dynamodb update-item \
    --table-name LikedSongs \
    --key '{ "Artist": {"S": "RMHighlander"}, "ReleaseDate": {"S": "2021-11-13"}}' \
    --update-expression "SET Genre = :newval" \
    --expression-attribute-values '{":newval":{"S":"Soft Rock"}}' \
    --return-values ALL_NEW \
    --endpoint-url=http://localhost:4566
```

* Get data:
```
aws dynamodb get-item --consistent-read \ 
    --table-name LikedSongs \
    --key '{ "Artist": {"S": "RMHighlander"}, "ReleaseDate": {"S": "2021-11-13"}}' \
    --endpoint-url=http://localhost:4566
```

* Query data:
(| LE | LT | GE | GT | BEGINS_WITH | BETWEEN |)
```
aws dynamodb query \
    --table-name LikedSongs \
    --key-condition-expression "Artist = :name" \
    --expression-attribute-values  '{":name":{"S":"RMHighlander"}}' \
    --endpoint-url=http://localhost:4566


aws dynamodb query \
    --table-name LikedSongs \
    --key-condition-expression "Artist = :a AND #date > :d" \
    --expression-attribute-values '{":a":{"S":"RMHighlander"}, ":d":{"S":"2022-01-01"}}' \
    --expression-attribute-names '{"#date": "ReleaseDate"}' \
    --endpoint-url=http://localhost:4566
```

read parameters:
```
--projection-expression: retrieve attributes
--filter-expression: filter items
--page-size: default 1000 items
--max-items: the number of items to show in the CLI (returns NextToken)
--starting-token: last NextToken to retrieve the next set of items
```

* Scan data:
```
aws dynamodb scan \
    --table-name LikedSongs \
    --endpoint-url=http://localhost:4566

aws dynamodb scan \
    --table-name LikedSongs \
    --projection-expression "Artist, Title" \
    --endpoint-url=http://localhost:4566

aws dynamodb scan \
    --table-name LikedSongs \
    --filter-expression "Artist = :a" \
    --expression-attribute-values '{":a": {"S":"RMHighlander"}}' \
    --endpoint-url=http://localhost:4566
```

page size will do several calls (is used to avoid timeouts):
```
aws dynamodb scan \
    --table-name LikedSongs \
    --page-size 1 \
    --endpoint-url=http://localhost:4566
```

max items (returns just the specified number of items):
```
aws dynamodb scan \
    --table-name LikedSongs \
    --max-items 1 \
    --endpoint-url=http://localhost:4566

aws dynamodb scan \
    --table-name LikedSongs \
    --max-items 1 \
    --starting-token <NextToken> \
    --endpoint-url=http://localhost:4566
```

* Batch read:
```
aws dynamodb batch-get-item \
    --request-items file://data-read.json \
    --return-consumed-capacity TOTAL \
    --endpoint-url=http://localhost:4566
```

### TTL
		TODO

## Resources

https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GettingStartedDynamoDB.html

https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.CoreComponents.html

https://docs.aws.amazon.com/cli/latest/reference/dynamodb/

https://aws.github.io/aws-sdk-go-v2/docs/

https://github.com/aws/aws-sdk-go-v2