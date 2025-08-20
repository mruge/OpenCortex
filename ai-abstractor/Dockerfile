FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ai-abstractor .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/ai-abstractor .

EXPOSE 8081

CMD ["./ai-abstractor"]