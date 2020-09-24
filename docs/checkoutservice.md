Checkout Service
===================

## Logs
To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/checkoutservice.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"go","service":"checkoutservice", "sourcecategory":"sourcecode", "file":"file"}]'
```

The log body contains valuable information (user Id, currency, transaction Id, ...).
We can extract those information with a Gork parser

```gork
rule1 order\s+confirmation\s+email\s+sent\s+to\s+\"%{data:email}\"

rule2 \[PlaceOrder\]\s+%{data::keyvalue("=","*\\[,\\]")}

rule3 payment\s+went\s+through\s+\(transaction_id\:\s+%{notSpace:transaction.id}\)
```

We also need to remap Date and Status fields with `timestamp` and `severity`

## APM

The first step in instrumenting a Go application is to install the Go Tracer.

Inside the `checkoutservice` directory run:

```sh
go get gopkg.in/DataDog/dd-trace-go.v1/ddtrace
```

Now we can instrument our application

### Traces

First, we create client interceptors to trace the different gRPB we are going to call, and server interceptors to trace our service

##### **`src/checkoutservice/main.go`**
```go
type checkoutService struct {
    ...
    si                    grpc.StreamClientInterceptor
    ui                    grpc.UnaryClientInterceptor
}

func main() {
    ...

    mustMapEnv(&svc.paymentSvcAddr, "PAYMENT_SERVICE_ADDR")

    svc.si = grpctrace.StreamClientInterceptor(grpctrace.WithServiceName(serviceName))
    svc.ui = grpctrace.UnaryClientInterceptor(grpctrace.WithServiceName(serviceName))

    // Create the server interceptor using the grpc trace package.
    ssi := grpctrace.StreamServerInterceptor(grpctrace.WithServiceName(serviceName))
    usi := grpctrace.UnaryServerInterceptor(grpctrace.WithServiceName(serviceName))
```

Then we add our interceptors to our server

##### **`src/checkoutservice/main.go`**
```go
srv = grpc.NewServer(grpc.StreamInterceptor(ssi), grpc.UnaryInterceptor(usi))
```

We also need to add our client interceptors to the servers we call

##### **`src/checkoutservice/main.go`**
```go
func (cs *checkoutService) quoteShipping(ctx context.Context, address *pb.Address, items []*pb.CartItem) (*pb.Money, error) {
    conn, err := grpc.DialContext(ctx, cs.shippingSvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
    
...

func (cs *checkoutService) getUserCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
    conn, err := grpc.DialContext(ctx, cs.cartSvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
    
...

func (cs *checkoutService) emptyUserCart(ctx context.Context, userID string) error {
    conn, err := grpc.DialContext(ctx, cs.cartSvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
    
...

func (cs *checkoutService) prepOrderItems(ctx context.Context, items []*pb.CartItem, userCurrency string) ([]*pb.OrderItem, error) {
	out := make([]*pb.OrderItem, len(items))

    conn, err := grpc.DialContext(ctx, cs.productCatalogSvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
    
...

func (cs *checkoutService) convertCurrency(ctx context.Context, from *pb.Money, toCurrency string) (*pb.Money, error) {
    conn, err := grpc.DialContext(ctx, cs.currencySvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
    
...

func (cs *checkoutService) chargeCard(ctx context.Context, amount *pb.Money, paymentInfo *pb.CreditCardInfo) (string, error) {
    conn, err := grpc.DialContext(ctx, cs.paymentSvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
    
...

func (cs *checkoutService) sendOrderConfirmation(ctx context.Context, email string, order *pb.OrderResult) error {
    conn, err := grpc.DialContext(ctx, cs.emailSvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
    
...

func (cs *checkoutService) shipOrder(ctx context.Context, address *pb.Address, items []*pb.CartItem) (string, error) {
	conn, err := grpc.DialContext(ctx, cs.shippingSvcAddr, grpc.WithInsecure(), grpc.WithStreamInterceptor(cs.si), grpc.WithUnaryInterceptor(cs.ui))
```

### Connecting log and traces

Datadog Go library doesn't support automatic log injection yet, so we need to inject it manually.

To simplify the process, we can extend the `StandardLogger` with a `WithSpan` function that will inject trace and span id in the log.

##### **`src/checkoutservice/logwrapper/logwrapper.go`**
```go
func (l *StandardLogger) WithSpan(span ddtrace.Span) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"dd.trace_id": span.Context().TraceID(),
		"dd.span_id":  span.Context().SpanID(),
	})
}
```

All is left to do is to grab span from context and call the `WithSpan` function.

##### **`src/checkoutservice/main.go`**
```go
func (cs *checkoutService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	span, _ := tracer.SpanFromContext(ctx)
    log.WithSpan(span).Infof("[PlaceOrder] user_id=%q user_currency=%q", req.UserId, req.UserCurrency)

    ...

    log.WithSpan(span).Infof("payment went through (transaction_id: %s)", txID)

    ...

    if err := cs.sendOrderConfirmation(ctx, req.Email, orderResult); err != nil {
		log.WithSpan(span).Warnf("failed to send order confirmation to %q: %+v", req.Email, err)
	} else {
		log.WithSpan(span).Infof("order confirmation email sent to %q", req.Email)
    }
    
    ...

}
```