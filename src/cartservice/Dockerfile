FROM mcr.microsoft.com/dotnet/core/sdk:3.1 AS build

WORKDIR /source

# copy csproj and restore as distinct layers
COPY *.csproj .
RUN dotnet restore -r linux-musl-x64

COPY . .
RUN dotnet publish -c release -o /app -r linux-musl-x64

FROM mcr.microsoft.com/dotnet/core/runtime:3.1-alpine
WORKDIR /app
COPY --from=build /app .
COPY nlog.config .

RUN GRPC_HEALTH_PROBE_VERSION=v0.3.2 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

RUN mkdir -p /opt/datadog && \
    wget -qO- https://github.com/DataDog/dd-trace-dotnet/releases/download/v1.18.0/datadog-dotnet-apm-1.18.0-musl.tar.gz \
    | tar xzf - -C /opt/datadog

RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.8/main' >> /etc/apk/repositories && \
    apk update --no-cache && \
    apk add --no-cache bash libc6-compat=1.1.19-r11

ENTRYPOINT ["./cartservice", "start"]