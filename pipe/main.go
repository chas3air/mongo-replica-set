package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// WRITE → PRIMARY (через HAProxy)
	writeURI := "mongodb://root:example@haproxy:27019/?authSource=admin"
	writeClient, err := mongo.Connect(ctx, options.Client().ApplyURI(writeURI))
	if err != nil {
		log.Fatal("write connect error:", err)
	}

	// READ → SECONDARY (через HAProxy)
	readURI := "mongodb://root:example@haproxy:27018/?authSource=admin&readPreference=secondary"
	readClient, err := mongo.Connect(ctx, options.Client().ApplyURI(readURI))
	if err != nil {
		log.Fatal("read connect error:", err)
	}

	collection := writeClient.Database("testdb").Collection("items")

	// Insert document
	doc := bson.M{
		"msg":  "Hello from Go!",
		"time": time.Now(),
	}

	insertResult, err := collection.InsertOne(ctx, doc)
	if err != nil {
		log.Fatal("insert error:", err)
	}

	fmt.Println("Inserted ID:", insertResult.InsertedID)

	// Wait for replication
	time.Sleep(1 * time.Second)

	// Read document
	var result bson.M
	err = readClient.Database("testdb").Collection("items").
		FindOne(ctx, bson.M{"_id": insertResult.InsertedID}).Decode(&result)
	if err != nil {
		log.Fatal("read error:", err)
	}

	fmt.Println("Read from secondary:", result)
}
