# syntax=docker/dockerfile:1
# Build
FROM --platform=linux/amd64 docker.io/golang:1.25-alpine AS builder
WORKDIR /workspace
COPY . .
# We must run with CGO_ENABLED=0 because otherwise the alpine container wont be able to launch it unless we install more packages
# We also must remove the "v" from the TARGETVARIAT since docker takes "v7" while go takes just "7"
RUN CGO_ENABLED=0 go build -o device-store-api

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /workspace/device-store-api ./
EXPOSE 8080
CMD ["./device-store-api"]