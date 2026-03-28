FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./.
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app -v ./.

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
RUN chmod +x /app
ENTRYPOINT /app