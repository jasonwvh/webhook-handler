# Webhook Handler

### Starting the services

```
docker compose up
```

### Using the service

```
curl localhost:8080/webhook -X POST -d '{"id":123,"url":"http://test.com"}'
```

or

```
curl localhost:8080/async-webhook -X POST -d '{"id":123,"url":"http://test.com"}'
```