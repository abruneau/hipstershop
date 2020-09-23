Cart Service
===================

## Logs
To start collecting logs, we need to annotate the pod definition to indicate to the Datadog agent to collect logs and their nature

##### **`kubernetes-manifests/cartservice.yaml`**
```yaml
annotations:
    ad.datadoghq.com/server.logs: '[{"source":"csharp","service":"cartservice", "sourcecategory":"sourcecode", "file":"file"}]'
```

The log body contains valuable information about the function called with parameters (user Id, product, quantity).
We can extract those information with a Gork parser

```regex
rule %{notSpace:action}\s+called with %{data::keyvalue("=","*\\[,\\]")}
```

## APM

The first step in implementing a C# application, is to install the .NET package.

##### **`src/cartservice/Dockerfile`**
```docker
...
FROM mcr.microsoft.com/dotnet/core/runtime:3.1-alpine
...
RUN mkdir -p /opt/datadog && \
    wget -qO- https://github.com/DataDog/dd-trace-dotnet/releases/download/v1.18.0/datadog-dotnet-apm-1.18.0-musl.tar.gz \
    | tar xzf - -C /opt/datadog

```

Next we need to add a few environment variables to the deployment

##### **`kubernetes-manifests/cartservice.yaml`**
```yaml
env:
...
    - name: DD_AGENT_HOST
        valueFrom:
        fieldRef:
            fieldPath: status.hostIP
    - name: DD_TRACE_AGENT_PORT
        value: "8126"
    - name: CORECLR_ENABLE_PROFILING
        value: "1"
    - name: CORECLR_PROFILER
        value: "{846F5F1C-F9AE-4B07-969E-05C26BC060D8}"
    - name: CORECLR_PROFILER_PATH
        value: "/opt/datadog/Datadog.Trace.ClrProfiler.Native.so"
    - name: DD_INTEGRATIONS
        value: "/opt/datadog/integrations.json"
    - name: DD_DOTNET_TRACER_HOME
        value: "/opt/datadog"
    - name: DD_LOGS_INJECTION
        value: "true"
    - name: DD_ENV
        value: "prod"
    - name: DD_SERVICE
        value: "cartservice"
    - name: DD_VERSION
        value: "1.0.0"
    - name: DD_TRACE_ENABLED
        value: "true"
    - name: DD_TRACE_DEBUG
        value: "false"
```

This application leverage GRPC framework that is not natively supported by the Datadog agent, so span information are limited.

To improve it, the first option is to use OpenTracing integration.

Add `Datadog.Trace.OpenTracing` and `OpenTracing.Contrib.Grpc` nugets packages to your project. Then add gRPC interceptor to your server

##### **`src/cartservice/Program.cs`**
```csharp
using OpenTracing.Contrib.Grpc.Interceptors;
using Datadog.Trace.OpenTracing;
...
static object StartServer(string host, int port, ICartStore cartStore){
    ...
    logger.Debug($"Trying to start a grpc server at  {host}:{port}");
    // Create an OpenTracing ITracer with the default setting
    OpenTracing.ITracer tracer = OpenTracingTracerFactory.CreateTracer();
    ServerTracingInterceptor tracingInterceptor = new ServerTracingInterceptor(tracer);
    ...
    Hipstershop.CartService.BindService(new CartServiceImpl(cartStore)).Intercept(tracingInterceptor),
```

The second option is to implement a gRPC interceptor with `Datadog.Trace` library. For the exercise, I only build a server unary interceptor.

##### **`src/cartservice/grpc_interceptor/server.cs`**
```csharp
namespace cartservice.grpc_interceptor
{
    public class DatadogInterceptor : Interceptor
    {
        private SpanContext GetTraceContextFromContext(ServerCallContext context) {
            var traceId = context.RequestHeaders.GetValue(HttpHeaderNames.TraceId);
            var parentId = context.RequestHeaders.GetValue(HttpHeaderNames.ParentId);

            if (traceId != null && parentId != null)
            {
                ulong traceIdNb = Convert.ToUInt64(traceId);
                ulong parentSpanId = Convert.ToUInt64(parentId);
                return new SpanContext(traceIdNb, parentSpanId);
            }

            return null;
        }

        public async override Task<TResponse> UnaryServerHandler<TRequest, TResponse>(
            TRequest request,
            ServerCallContext context,
            UnaryServerMethod<TRequest, TResponse> continuation)
        {
            
            using (var scope = Tracer.Instance.StartActive("grpc.server", parent: GetTraceContextFromContext(context)))
            {
                var span = scope.Span;
                span.Type = SpanTypes.Custom;
                span.ResourceName = context.Method;
                try
                {
                    return await continuation(request, context);
                }
                catch (Exception e)
                {
                    span.SetException(e);
                    throw new RpcException(Status.DefaultCancelled, e.Message);
                }
            }
        }
    }
}
```