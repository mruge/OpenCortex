FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o data-abstractor .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/data-abstractor .

EXPOSE 8080

CMD ["./data-abstractor"]