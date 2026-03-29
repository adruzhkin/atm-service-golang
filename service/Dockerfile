FROM golang:1.26.1-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o main cmd/main.go

FROM alpine:3.21
EXPOSE 5000

COPY --from=builder /build/main /main

CMD ["/main"]
