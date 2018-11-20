FROM golang:1.11 AS builder

RUN curl https://raw.githubusercontent.com/golang/dep/v0.5.0/install.sh | sh

WORKDIR $GOPATH/src/github.com/glutamatt/webstun
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY server ./server
COPY client ./client
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /webstun .

FROM alpine:3.8
WORKDIR /
COPY --from=builder /webstun /bin
ARG server_port=443
ENV PORT=$server_port
CMD ["/bin/sh", "-c", "webstun -run server -port ${PORT} -insecure"]
