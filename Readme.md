# Simple hello world app to test rabbitmq connection

Required environment variables:
```shell
export RABBITMQ_DEFAULT_USER="guest"
export RABBITMQ_DEFAULT_PASS="guest"
export RABBITMQ_SERVER="localhost:5672"
export RABBITMQ_DEFAULT_VHOST="/"
```

optional, doesn't work on docker:
```shell
export PORT="8008"
```

Run using docker:
```shell
docker run -e RABBITMQ_DEFAULT_USER -e RABBITMQ_DEFAULT_PASS -e RABBITMQ_SERVER -e RABBITMQ_DEFAULT_VHOST joaomorenorf/hello-rabbitmq-app:1.0.0
```

Run directly:
```shell
go run main.go
```

Run a rabbitmq with the same variables:
```shell
docker run -e RABBITMQ_DEFAULT_USER -e RABBITMQ_DEFAULT_PASS -e RABBITMQ_SERVER -e RABBITMQ_DEFAULT_VHOST -d --hostname rabbit --name docker-rabbit -p 5672:5672 rabbitmq:3
```