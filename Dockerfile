FROM golang:1.19.3 as builder
ENV GOOS linux
ENV CGO_ENABLED 0
RUN mkdir appbin
WORKDIR /app 
COPY go.mod go.sum ./ 
RUN go mod download
COPY . .
RUN go build -o /appbin/producer ./producer
RUN go build -o /appbin/consumer ./consumer 
RUN go build -o /appbin/http_tester ./http_tester 
WORKDIR /


FROM alpine:3.14 as production
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /appbin/* .
