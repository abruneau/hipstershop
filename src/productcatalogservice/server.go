package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	pb "github.com/abruneau/hipstershop/src/productcatalogservice/genproto"
	"github.com/abruneau/hipstershop/src/productcatalogservice/logwrapper"
	"github.com/abruneau/hipstershop/src/productcatalogservice/store"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	grpctrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/google.golang.org/grpc"
)

var (
	log          *logwrapper.StandardLogger
	extraLatency time.Duration
	serviceName  = "productcatalogservice"

	port = "3550"
)

func init() {
	log = logwrapper.NewLogger()
	log.Level = logrus.InfoLevel
	log.Out = os.Stdout
}

func main() {
	tracer.Start()
	defer tracer.Stop()
	flag.Parse()

	// set injected latency
	if s := os.Getenv("EXTRA_LATENCY"); s != "" {
		v, err := time.ParseDuration(s)
		if err != nil {
			log.Fatalf("failed to parse EXTRA_LATENCY (%s) as time.Duration: %+v", v, err)
		}
		extraLatency = v
		log.Infof("extra latency enabled (duration: %v)", extraLatency)
	} else {
		extraLatency = time.Duration(0)
	}

	err := profiler.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer profiler.Stop()

	catalog, err := store.NewMogoStore(log)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	ctx := context.Background()
	defer catalog.Disconnect(ctx)

	if err = catalog.LoadCatalog(ctx); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.Infof("starting grpc server at :%s", port)
	run(port, catalog)
	select {}
}

func run(port string, catalog store.Store) string {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	// Create the server interceptor using the grpc trace package.
	si := grpctrace.StreamServerInterceptor(grpctrace.WithServiceName(serviceName))
	ui := grpctrace.UnaryServerInterceptor(grpctrace.WithServiceName(serviceName))

	var srv *grpc.Server
	srv = grpc.NewServer(grpc.StreamInterceptor(si), grpc.UnaryInterceptor(ui))

	svc := &productCatalog{
		catalog: catalog,
	}

	pb.RegisterProductCatalogServiceServer(srv, svc)
	healthpb.RegisterHealthServer(srv, svc)
	go srv.Serve(l)
	return l.Addr().String()
}

type productCatalog struct {
	catalog store.Store
}

func (p *productCatalog) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (p *productCatalog) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (p *productCatalog) ListProducts(ctx context.Context, _ *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(extraLatency)
	products, err := p.catalog.List(ctx)
	return &pb.ListProductsResponse{Products: products}, err
}

func (p *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(extraLatency)

	found, err := p.catalog.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
	}
	return found, nil
}

func (p *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(extraLatency)

	products, err := p.catalog.Find(ctx, req.Query)
	return &pb.SearchProductsResponse{Results: products}, err
}
