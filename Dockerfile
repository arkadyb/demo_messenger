########################
######## Build #########
########################

FROM golang:1.11 AS build

COPY . /demo_messenger
WORKDIR /demo_messenger

RUN make build

########################
####### Publish ########
########################

FROM alpine:3.7

ARG SERVICE

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=build /demo_messenger/dist/demo_messenger app

CMD ["/app"]
