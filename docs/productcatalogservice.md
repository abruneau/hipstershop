Product Catalog Service
=========================

## Logs

To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/productcatalogservice.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"go","service":"productcatalogservice", "sourcecategory":"sourcecode", "file":"file"}]'
```

## APM

The first step in instrumenting a Go application is to install the Go Tracer.

Inside the `productcatalogservice` directory run:

```sh
go get gopkg.in/DataDog/dd-trace-go.v1/ddtrace
```

Now we can instrument our application

### Traces

First, we create server interceptors, and inject them in our server to trace our service 

##### **`src/productcatalogservice/server.go`**
```go
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
```

We also need to instrument our mongo integration

##### **`src/productcatalogservice/store/store.go`**
```go
// NewMogoStore initialize a new Mongodb connexion
func NewMogoStore(log *logwrapper.StandardLogger) (Store, error) {
	...

	if mongoURL := os.Getenv("MONGO_URL"); mongoURL != "" {
		opts := options.Client()
		opts.Monitor = mongotrace.NewMonitor(mongotrace.WithAnalytics(true), mongotrace.WithServiceName("mongo"))
		opts.ApplyURI(mongoURL)
		client, err := mongo.Connect(ctx, opts)
```

Finally we can instrument our functions a little bit more

##### **`src/productcatalogservice/store/mongodb.go`**
```go
// LoadCatalog load catalog from file
func (m *mongodb) LoadCatalog(ctx context.Context) error {
	span, _ := tracer.StartSpanFromContext(ctx, "store.mongodb.LoadCatalog")
    defer span.Finish()
    
...

func (m *mongodb) List(ctx context.Context) (products []*pb.Product, err error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "store.mongodb.List")
    defer span.Finish()

...

func (m *mongodb) Get(ctx context.Context, id string) (product *pb.Product, err error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "store.mongodb.Get")
    defer span.Finish()
    
...

func (m *mongodb) Find(ctx context.Context, text string) (products []*pb.Product, err error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "store.mongodb.Find")
	defer span.Finish()
```

Next we need to add a few environment variables to the deployment

##### **`kubernetes-manifests/productcatalogservice.yaml`**
```yaml
- name: DD_AGENT_HOST
    valueFrom:
    fieldRef:
        fieldPath: status.hostIP
- name: DD_ENV
    value: "prod"
- name: DD_SERVICE
    value: "productcatalogservice"
- name: DD_VERSION
    value: "1.0.0"
```

### Connecting log and traces

Datadog Go library doesn't support automatic log injection yet, so we need to inject it manually.

To simplify the process, we can extend the `StandardLogger` with a `WithSpan` function that will inject trace and span id in the log.

##### **`src/productcatalogservice/logwrapper/logwrapper.go`**
```go
func (l *StandardLogger) WithSpan(span ddtrace.Span) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
	})
}
```

All is left to do is to grab span from context and call the `WithSpan` function.

##### **`src/productcatalogservice/store/store.go`**
```go
// LoadCatalog load catalog from file
func (m *mongodb) LoadCatalog(ctx context.Context) error {
    ...
    
	if len(products) == 0 {
        m.log.WithSpan(span).Info("Loading catalog")
        
        ...
        
		m.log.WithSpan(span).Info("Catalog loaded")
	} else {
		m.log.WithSpan(span).Info("Catalog already loaded")
	}
	return nil
}

```

### Profiling

##### **`src/productcatalogservice/server.go`**
```go
func main() {
	tracer.Start()
	defer tracer.Stop()
    
    ...

	err := profiler.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer profiler.Stop()
```