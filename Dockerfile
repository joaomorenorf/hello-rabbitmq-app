FROM golang:1.22.1 as builder
WORKDIR /app
COPY main.go ./
COPY go.mod ./
COPY go.sum ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /hello-rabbitmq-app

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /hello-rabbitmq-app /hello-rabbitmq-app

ENV PORT 8008
ARG RABBITMQ_DEFAULT_USER
ARG RABBITMQ_DEFAULT_PASS
ARG RABBITMQ_SERVER
ARG RABBITMQ_DEFAULT_VHOST

USER nonroot:nonroot
EXPOSE 8008
CMD ["/hello-rabbitmq-app"]
