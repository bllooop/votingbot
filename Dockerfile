FROM golang:1.24

WORKDIR /app
RUN go version
ENV $GOPATH=/

COPY . .

RUN go mod download
RUN go build -o votingbot ./cmd/main.go

EXPOSE 8080

CMD ["./votingbot"]