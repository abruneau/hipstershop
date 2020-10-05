package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	pb "github.com/abruneau/hipstershop/src/productcatalogservice/genproto"
	"github.com/abruneau/hipstershop/src/productcatalogservice/logwrapper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

// Store interface
type Store interface {
	Find(context.Context, string) ([]*pb.Product, error)
	Get(context.Context, string) (*pb.Product, error)
	List(context.Context) ([]*pb.Product, error)
	LoadCatalog(context.Context) error
	Disconnect(context.Context)
}

// NewMogoStore initialize a new Mongodb connexion
func NewMogoStore(log *logwrapper.StandardLogger) (Store, error) {
	ctx := context.Background()
	var m = &mongodb{
		log: log,
	}

	if mongoURL := os.Getenv("MONGO_URL"); mongoURL != "" {
		opts := options.Client()
		opts.Monitor = mongotrace.NewMonitor(mongotrace.WithAnalytics(true), mongotrace.WithServiceName("mongo"))
		opts.ApplyURI(mongoURL)
		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			return nil, err
		}
		m.client = client

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
