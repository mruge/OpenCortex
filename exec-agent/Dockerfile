FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o exec-agent .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/exec-agent .

# Create work directory for container execution
RUN mkdir -p /tmp/exec-agent

EXPOSE 8082 9000

CMD ["./exec-agent"]