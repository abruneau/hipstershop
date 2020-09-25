Frontend
==================

## Logs
To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/frontend.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"go","service":"frontend", "sourcecategory":"sourcecode", "file":"file"}]'
```

## APM

The first step in instrumenting a Go application is to install the Go Tracer.

Inside the `frontend` directory run:

```sh
go get gopkg.in/DataDog/dd-trace-go.v1/ddtrace
```

Now we can instrument our application

### Mux

Our application implement a mux server, so we are going to replace the standard mux library with Datadog implementation

##### **`src/frontend/main.go`**
```go
import (
    ...
    "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
```

And then start a tracer

##### **`src/frontend/main.go`**
```go
func main() {
	tracer.Start(tracer.WithAnalytics(true))
    defer tracer.Stop()
    
    ...

    r := mux.NewRouter(mux.WithServiceName("frontend"))
```

### gRPC

Our frontend service calls some gRPC services we want to track. So we need to create interceptors.

##### **`src/frontend/main.go`**
```go
func mustConnGRPC(ctx context.Context, conn **grpc.ClientConn, addr string) {
	var err error
	si := grpctrace.StreamClientInterceptor(grpctrace.WithServiceName(serviceName))
	ui := grpctrace.UnaryClientInterceptor(grpctrace.WithServiceName(serviceName))
	*conn, err = grpc.DialContext(ctx, addr,
		grpc.WithInsecure(),
		grpc.WithTimeout(time.Second*3),
		grpc.WithStreamInterceptor(si),
		grpc.WithUnaryInterceptor(ui),
	)
	...
```

### Profiling

We can unable continuous profiling for our traces

##### **`src/frontend/main.go`**
```go
func main() {
    ...
    log.Out = os.Stdout

    err := profiler.Start()
    if err != nil {
        log.Fatal(err)
    }
    defer profiler.Stop()

    ...
```

### Connecting log and traces

We need to retrieve the span and trace id from the context and inject it in the logs

##### **`src/frontend/middleware.go`**
```go
func (lh *logHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
    span, _ := tracer.SpanFromContext(ctx)
    ...
    log := lh.log.WithFields(logrus.Fields{
		"http.req.path":   r.URL.Path,
		"http.req.method": r.Method,
		"http.req.id":     requestID.String(),
	}).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })
```

##### **`src/frontend/handlers.go`**
```go
func (fe *frontendServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })
    
...

func (fe *frontendServer) productHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })
    
...

func (fe *frontendServer) addToCartHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })
    
...

func (fe *frontendServer) emptyCartHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })

...

func (fe *frontendServer) viewCartHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })

...

func (fe *frontendServer) placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })
    
...

func (fe *frontendServer) logoutHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })
    
...

func (fe *frontendServer) setCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger).WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
    })
    
...
```

## RUM

We can also monitor real users by adding a bundle to our page header

##### **`src/frontend/templates/header.html`**
```html
<head>
    ...
    <script src="https://www.datadoghq-browser-agent.com/datadog-rum.js" type="text/javascript">
    </script>
    <script>
        window.DD_RUM && window.DD_RUM.init({
            applicationId: <APP_ID>,
            clientToken: <CLIENT_TOKEN>,
            site: 'datadoghq.com',
            sampleRate: 100,
            trackInteractions: true
        });
    </script>
</head>
```