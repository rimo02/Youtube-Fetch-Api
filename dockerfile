FROM golang:1.23.2-bullseye
WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o main .

EXPOSE 3000

CMD ["./main"]