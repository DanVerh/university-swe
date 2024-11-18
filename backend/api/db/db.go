// db.go
package db

import (
    "log"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo server uri
const dbUri = "mongodb://localhost:27017"

type Database struct {
	Client *mongo.Client
}

func DbConnect() *Database {
	db := &Database{
		Client: connectClient(),
	}

	return db
}

func connectClient() *mongo.Client {
    client, err := mongo.NewClient(options.Client().ApplyURI(dbUri))
    if err != nil {
        log.Fatal(err)
    }

    err = client.Connect(nil)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Connected to MongoDB")
    return client
}


// Close disconnects the MongoDB client and drops the connection
func (db *Database) DbDisconnect() {
	// Disconnect the MongoDB client without using context
	err := db.Client.Disconnect(nil)
	if err != nil {
		log.Fatalf("Failed to disconnect MongoDB client: %v", err)
	}

	log.Println("Disconnected from MongoDB")
}