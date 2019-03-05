# demo_messenger

Demo messenger is REST SMS delivery service showcase.

Service implements rate limiting and circuit breaker patterns. Uses Postgres as persistent data storage and Redis on rate limiter site.

## How to run

Service depends on persistant storage (Postgres), cache (Redis) and Message Bird client (used for sms delivery). There are two options to run it - run both (Redis and Postgres) on your local machine and run service with `make run` command or use Docker Compose.
In case of `make run`, please care to update/create `.env` file with the host names and passwords. Also, dont forget to get your Message Bird access key and put it into respective env variable in `.env` file.
```.env
LISTEN_PORT=8085

BUFFER_DB_CONNECTION_STRING=postgres://postgres@localhost:5432/postgres?sslmode=disable

REDIS_HOST=localhost:6379
REDIS_PWD=123456

MESSAGE_BIRD_KEY=[Your-MessageBird-Key]
```
To run service with `docker-compose`, update `docker-compose.yml` file with Message Bird access key you got and run `docker-compose up -d` command. It should run 3 containers - postgres, redis and demo_messenger.
When containers are ready (give it at least 5 seconds to start pg and redis), you should be able to call service's health endpoint at `http://localhost:8085/health` and get `true` in response confirming service is up and running. 

## How to use

Service exposes multiple endpoints:
- GET `/health` - health endpoint
- GET `/metrics` - Prometheus metrics endpoint
- POST `/v1/send/sms` - sms delivery via Message Bird endpoint
 
POST message body format (JSON):
```json
{
	"recipient": "PhoneNumber",
	"originator": "UniqueName OR PhoneNumber",
	"message": "Message"
}
```
Note: all fields are required and message body cant be longer than 160 characters.

Service is setup to batch delivery requests from same originator with same message body, therefore in case of multiple requests would be recorded to deliver sms notifications from originator `abc` to phone numbers `1`, `2` and `3` with message `hello world` all of them will be send together.
Batched messages are being send every second. Queued messages are delivered on first-in-first-out fashion.

## License
 
The MIT License (MIT)
