package store

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	pb "github.com/abruneau/hipstershop/src/productcatalogservice/genproto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongodb struct {
	client  *mongo.Client
	catalog *mongo.Collection
	log     *logrus.Logger
}

// Disconnect disconnect us from the database
func (m *mongodb) Disconnect() {
	ctx := context.Background()
	m.client.Disconnect(ctx)
}

// LoadCatalog load catalog from file
func (m *mongodb) LoadCatalog() error {
	ctx := context.Background()
	products, err := m.List()
	if err != nil {
		return err
	}

	if len(products) == 0 {
		m.log.Info("Loading catalog")
		var catalog pb.ListProductsResponse
		catalogJSON, err := ioutil.ReadFile("products.json")
		if err != nil {
			return fmt.Errorf("failed to open product catalog json file: %v", err)
		}
		if err := jsonpb.Unmarshal(bytes.NewReader(catalogJSON), &catalog); err != nil {
			return fmt.Errorf("failed to parse the catalog JSON: %v", err)
		}
		for i := range catalog.Products {
			doc := catalog.Products[i]
			_, insertErr := m.catalog.InsertOne(ctx, doc)
			if insertErr != nil {
				return fmt.Errorf("insertOne ERROR: %v", insertErr)
			}
		}
		m.log.Info("Catalog loaded")
	} else {
		m.log.Info("Catalog already loaded")
	}
	return nil
}

// List lists products
func (m *mongodb) List() (products []*pb.Product, err error) {
	ctx := context.Background()
	cursor, err := m.catalog.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

// Get gets a prodict from ID
func (m *mongodb) Get(id string) (product *pb.Product, err error) {
	err = m.catalog.FindOne(context.Background(), bson.M{"id": id}).Decode(&product)
	return
}

// Find searches products based on a string
func (m *mongodb) Find(text string) (products []*pb.Product, err error) {
	ctx := context.Background()
	query := bson.M{
		"$text": bson.M{
			"$search": text,
		},
	}

	cursor, err := m.catalog.Find(ctx, query)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return
}
