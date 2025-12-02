FROM golang:alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init restreamer && go build -o server main.go

FROM alpine:latest
RUN apk add --no-cache ffmpeg nginx ca-certificates gettext \
    && mkdir -p /run/nginx \
    && mkdir -p /dev/shm/hls

WORKDIR /app
COPY --from=builder /app/server /app/server
COPY nginx.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 80
CMD ["/app/entrypoint.sh"]
