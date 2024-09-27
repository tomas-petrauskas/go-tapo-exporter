# Build app
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tapo-exporter .

# Final image
FROM alpine:edge
WORKDIR /app
COPY --from=build /app/tapo-exporter .
RUN apk --no-cache add ca-certificates tzdata
EXPOSE 8086
ENTRYPOINT ["/app/tapo-exporter"]