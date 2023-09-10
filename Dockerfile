FROM golang:1.20 AS development
WORKDIR /node
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN go mod download
CMD go run main.go

FROM golang:alpine AS builder
WORKDIR /node
COPY . .
RUN go build -o /go/bin/node ./main.go

FROM alpine:latest AS production
COPY --from=builder /go/bin/node /go/bin/node
ENTRYPOINT ["/go/bin/node"]
