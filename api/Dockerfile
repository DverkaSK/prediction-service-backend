FROM golang:1.22

WORKDIR /app

COPY go.mod ./

RUN go mod download && go mod tidy

COPY . .

WORKDIR /app/api
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]