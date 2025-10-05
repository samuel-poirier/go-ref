# Stage 1: Base image with Go and Air installed
FROM golang:1.25-alpine AS build

# Set working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./publisherService ./publisherService

WORKDIR /app/publisherService/cmd/publisher

RUN go build

FROM scratch AS final

WORKDIR /app

COPY --from=build /app/publisherService/cmd/publisher .

ENTRYPOINT ["/app/publisher"]
