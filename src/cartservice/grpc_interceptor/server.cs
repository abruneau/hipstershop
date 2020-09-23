using Grpc.Core;
using Grpc.Core.Interceptors;
using System.Threading.Tasks;
using System;
using Datadog.Trace;

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