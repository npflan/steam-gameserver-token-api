FROM golang:1.17.1-alpine AS builder
RUN mkdir -p /go/src/github.com/npflan/steam-gameserver-token-api
WORKDIR /go/src/github.com/npflan/steam-gameserver-token-api
COPY . .
RUN go get -d . && \
    CGO_ENABLED=0 GOOS=linux go build -a -o steam-gameserver-token-api .

FROM alpine:3.14
RUN addgroup -g 1000 -S go && \
    adduser -u 1000 -S web -G go && \
    apk add --no-cache ca-certificates tzdata
WORKDIR /home/web
COPY --from=builder /go/src/github.com/npflan/steam-gameserver-token-api/steam-gameserver-token-api /home/web
EXPOSE 8000

USER web

CMD ["./steam-gameserver-token-api"]
