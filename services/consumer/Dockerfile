# Stage 1: Base image with Go and Air installed
FROM golang:1.25-alpine AS build

# Set working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./consumerService ./consumerService

WORKDIR /app/consumerService/cmd/consumer

RUN go build

FROM scratch AS final

WORKDIR /app

COPY --from=build /app/consumerService/cmd/consumer .

ENTRYPOINT ["/app/consumer"]
