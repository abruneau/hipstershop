FROM golang:1.14-alpine AS mod
RUN apk add git

ENV PROJECT frontend
WORKDIR $GOPATH/$PROJECT

COPY go.* ./
RUN GO111MODULE=on go mod download

FROM golang:1.14-alpine AS build
COPY --from=mod $GOCACHE $GOCACHE
COPY --from=mod $GOPATH/pkg/mod $GOPATH/pkg/mod
ENV PROJECT frontend
WORKDIR $GOPATH/$PROJECT

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o /frontend .

FROM scratch AS release
WORKDIR /frontend
COPY --from=build /frontend ./server
COPY ./templates ./templates
COPY ./static ./static
EXPOSE 8080
ENTRYPOINT ["/frontend/server"]
