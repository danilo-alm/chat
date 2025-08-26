# 1. Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

ENV PATH="$PATH:$(go env GOPATH)/bin"

COPY last-seen-service/go.mod last-seen-service/go.sum ./
RUN go mod download

COPY ./protos/ ./protos/
RUN mkdir -p pb && \
    protoc -I /usr/include -I protos \
        --go_out=./pb --go_opt=paths=source_relative \
		--go-grpc_out=./pb --go-grpc_opt=paths=source_relative \
	    lastseen.proto

COPY last-seen-service/main.go ./main.go

RUN go build -o myapp main.go

# 2. Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/myapp .

EXPOSE 8080

CMD ["./myapp"]
