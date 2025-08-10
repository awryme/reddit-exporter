FROM golang:1.24 as builder
COPY ./ /go/src/reddit-exporter
WORKDIR /go/src/reddit-exporter
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go build -v -o /app/server ./cmd/server/

FROM alpine/openssl as openssl
RUN openssl s_client -showcerts -connect www.reddit.com:443 </dev/null > /etc/reddit-cert.crt
RUN openssl s_client -showcerts -connect telegram.org:443 </dev/null > /etc/telegram-cert.crt
RUN ls /etc/*.crt
RUN cat /proc/sys/kernel/random/uuid | tr -d - > /etc/machine-id
RUN cat /etc/machine-id

FROM alpine as runner
COPY --from=openssl /etc/*.crt /etc/ssl/certs/
COPY --from=openssl /etc/machine-id /etc/machine-id 
COPY --from=builder /app/server /app/server
ENV DIR = /app/http_books
ENV BASIC_DIR = /app/books
ENTRYPOINT [ "/app/server" ]
