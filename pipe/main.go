package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	uri := "mongodb://localhost:27018/?readPreference=secondary&directConnection=true"

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	fmt.Println("Testing HAProxy load balancing...")

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var result map[string]interface{}
		err := client.Database("admin").RunCommand(ctx, map[string]interface{}{"hello": 1}).Decode(&result)
		if err != nil {
			log.Println("Error:", err)
			continue
		}

		// поле "me" содержит адрес ноды, которая ответила
		fmt.Printf("Request %d → answered by: %v\n", i+1, result["me"])

		time.Sleep(500 * time.Millisecond)
	}
}
