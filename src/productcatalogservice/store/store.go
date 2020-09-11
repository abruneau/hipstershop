package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	pb "github.com/abruneau/hipstershop/src/productcatalogservice/genproto"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// Store interface
type Store interface {
	Find(string) ([]*pb.Product, error)
	Get(string) (*pb.Product, error)
	List() ([]*pb.Product, error)
	LoadCatalog() error
	Disconnect()
}

// NewMogoStore initialize a new Mongodb connexion
func NewMogoStore(log *logrus.Logger) (Store, error) {
	ctx := context.Background()
	var m = &mongodb{
		log: log,
	}

	if mongoURL := os.Getenv("MONGO_URL"); mongoURL != "" {
		client, err := mongo.NewClient(options.Client().ApplyURI(mongoURL))
		if err != nil {
			return nil, err
		}
		m.client = client
		err = client.Connect(ctx)
		if err != nil {
			return nil, err
		}

		storeDatabase := client.Database("store")
		m.catalog = storeDatabase.Collection("products")

		log.Info("Connected to the store")
		index := mongo.IndexModel{
			Keys: bsonx.Doc{{Key: "Name", Value: bsonx.String("text")}},
		}

		_, err = m.catalog.Indexes().CreateOne(ctx, index)
		if err != nil {
			return nil, fmt.Errorf("Indexes().CreateOne() ERROR: %v", err)
		}
		return m, nil
	}
	return nil, errors.New("failed to get MONGO_URL from env")
}
