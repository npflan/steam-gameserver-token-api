FROM golang:1.11.0 AS base
RUN mkdir -p /go/src/github.com/npflan/steam-gameserver-token-api
WORKDIR /go/src/github.com/npflan/steam-gameserver-token-api
COPY . .
RUN apt -y update && apt -y install musl-tools ca-certificates
RUN go get -d . && \
    CC=$(which musl-gcc) go build --ldflags '-w -linkmode external -extldflags "-static"' .

FROM alpine:3.7
RUN addgroup -g 1000 -S go && \
    adduser -u 1000 -S username -G go && \
    apk add --no-cache ca-certificates tzdata
WORKDIR /home/go
COPY --from=base /go/src/github.com/npflan/steam-gameserver-token-api/steam-gameserver-token-api /home/go
EXPOSE 80
CMD ["./steam-gameserver-token-api"]
