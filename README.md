# EMR blockchain project

A blockchain-based Electronic Medical Record management, with roles permissions.

## Tutorial

This project was started following [this tutorial](https://www.codementor.io/codehakase/building-a-simple-blockchain-with-go-k7crur06v).

## Usage

### Run app

```shell
$ go run blockchain.go

```

### Create book

```json
POST http://localhost:3000/new

body: {
	"title": "Book",
	"author":"JohnDoe",
	"isbn":"909090",
	"publish_date":"2018-02-01"
}
```

### Create block

```json
POST http://localhost:3000

body:
{
	"book_id": "id_generated_with_previous_request",
	"user": "john Doe",
	"checkout_date":"2018-01-01"
}
```

### Get blockchain

GET `http://localhost:3000`
