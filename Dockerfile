# Build stage
FROM golang:1.22 as builder
WORKDIR /app
COPY ./main.go .
COPY ./go.mod .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloudflare-ddns-updater .

# Final stage
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/cloudflare-ddns-updater .
CMD ["./cloudflare-ddns-updater"]
