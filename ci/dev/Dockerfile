FROM golang:1.16.5-alpine3.13

WORKDIR /app

COPY . /app/

RUN go build -o /timeline

CMD ["/timeline"]
