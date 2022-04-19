FROM golang:1.17
EXPOSE 5000

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go mod download
RUN go build -o cmd/main cmd/main.go

CMD ["cmd/main"]