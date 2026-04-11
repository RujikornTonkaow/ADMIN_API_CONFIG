package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Connect(ctx context.Context, uri, dbName string, log *slog.Logger) (*mongo.Database, func(), error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	if err := client.Ping(connectCtx, readpref.Primary()); err != nil {
		return nil, nil, fmt.Errorf("pinging MongoDB: %w", err)
	}

	log.Info("connected to MongoDB", "database", dbName)

	disconnect := func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(disconnectCtx); err != nil {
			log.Error("disconnecting from MongoDB", "error", err)
		}
		log.Info("disconnected from MongoDB")
	}

	return client.Database(dbName), disconnect, nil
}
