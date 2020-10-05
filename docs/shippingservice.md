Shipping Service
=========================

## Logs

To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/productcatalogservice.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"go","service":"shippingservice", "sourcecategory":"sourcecode", "file":"file"}]'
```

## APM

The first step in instrumenting a Go application is to install the Go Tracer.

Inside the `shippingservice` directory run:

```sh
go get gopkg.in/DataDog/dd-trace-go.v1/ddtrace
```

Now we can instrument our application

### Traces

First, we create server interceptors, and inject them in our server to trace our service 

##### **`src/shippingservice/main.go`**
```go
func main() {
    ...

	// Create the server interceptor using the grpc trace package.
	si := grpctrace.StreamServerInterceptor(grpctrace.WithServiceName(serviceName))
	ui := grpctrace.UnaryServerInterceptor(grpctrace.WithServiceName(serviceName))

	var srv *grpc.Server
	srv = grpc.NewServer(grpc.StreamInterceptor(si), grpc.UnaryInterceptor(ui))
```

Then we can instrument our functions a little bit more

##### **`src/shippingservice/quote.go`**
```go
// CreateQuoteFromCount takes a number of items and returns a Price struct.
func CreateQuoteFromCount(ctx context.Context, count int) Quote {
	span, ctx := tracer.StartSpanFromContext(ctx, "CreateQuoteFromCount")
	defer span.Finish()
	return CreateQuoteFromFloat(ctx, quoteByCountFloat(ctx, count))
}

// CreateQuoteFromFloat takes a price represented as a float and creates a Price struct.
func CreateQuoteFromFloat(ctx context.Context, value float64) Quote {
	span, ctx := tracer.StartSpanFromContext(ctx, "CreateQuoteFromFloat")
	defer span.Finish()
	units, fraction := math.Modf(value)
	return Quote{
		uint32(units),
		uint32(math.Trunc(fraction * 100)),
	}
}

// quoteByCountFloat takes a number of items and generates a price quote represented as a float.
func quoteByCountFloat(ctx context.Context, count int) float64 {
	span, ctx := tracer.StartSpanFromContext(ctx, "quoteByCountFloat")
	defer span.Finish()
	if count == 0 {
		return 0
	}
	count64 := float64(count)
	var p = 1 + (count64 * 0.2)
	return count64 + math.Pow(3, p)
}
```

##### **`src/shippingservice/tracker.go`**
```go
// CreateTrackingID generates a tracking ID.
func CreateTrackingID(ctx context.Context, salt string) string {
	span, ctx := tracer.StartSpanFromContext(ctx, "CreateTrackingID")
	defer span.Finish()
	...
}

// getRandomLetterCode generates a code point value for a capital letter.
func getRandomLetterCode(ctx context.Context) uint32 {
	span, ctx := tracer.StartSpanFromContext(ctx, "getRandomLetterCode")
	defer span.Finish()
	return 65 + uint32(rand.Intn(25))
}

// getRandomNumber generates a string representation of a number with the requested number of digits.
func getRandomNumber(ctx context.Context, digits int) string {
	span, ctx := tracer.StartSpanFromContext(ctx, "getRandomNumber")
	defer span.Finish()
	...
}
```

Finally we need to add a few environment variables to the deployment

##### **`kubernetes-manifests/shippingservice.yaml`**
```yaml
- name: DD_AGENT_HOST
    valueFrom:
    fieldRef:
        fieldPath: status.hostIP
- name: DD_ENV
    value: "prod"
- name: DD_SERVICE
    value: "shippingservice"
- name: DD_VERSION
    value: "1.0.0"
```

### Connecting log and traces

Datadog Go library doesn't support automatic log injection yet, so we need to inject it manually.

To simplify the process, we can extend the `StandardLogger` with a `WithSpan` function that will inject trace and span id in the log.

##### **`src/shippingservice/logwrapper/logwrapper.go`**
```go
func (l *StandardLogger) WithSpan(span ddtrace.Span) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
	})
}
```

All is left to do is to grab span from context and call the `WithSpan` function.

##### **`src/shippingservice/main.go`**
```go
// GetQuote produces a shipping quote (cost) in USD.
func (s *server) GetQuote(ctx context.Context, in *pb.GetQuoteRequest) (*pb.GetQuoteResponse, error) {
	span, _ := tracer.SpanFromContext(ctx)
	log.WithSpan(span).Info("[GetQuote] received request")
    defer log.WithSpan(span).Info("[GetQuote] completed request")
    
...

func (s *server) ShipOrder(ctx context.Context, in *pb.ShipOrderRequest) (*pb.ShipOrderResponse, error) {
	span, _ := tracer.SpanFromContext(ctx)
	log.WithSpan(span).Info("[ShipOrder] received request")
    defer log.WithSpan(span).Info("[ShipOrder] completed request")
```

### Profiling

##### **`src/shippingservice/main.go`**
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