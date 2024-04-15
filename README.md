# Webhook Handler

### Starting the services

```
docker compose up --build
```

### Using the service

```
curl localhost:8080/webhook -X POST -d '{"id":101,"url":"http://google.com","seq":1}'
```

or

```
curl localhost:8080/async-webhook -X POST -d '{"id":201,"url":"http://google.com"."seq":1}'
```