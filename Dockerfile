FROM node:14.15-alpine AS nodebuild

WORKDIR /workdir
COPY web/ /workdir

RUN yarn install && yarn build

FROM golang:1.14-alpine as build

WORKDIR /go/src/super
ADD . /go/src/super

RUN go get -d -v ./...

RUN go build -o /go/bin/super

FROM alpine:3.9
WORKDIR /deployment
COPY --from=build /go/bin/super /deployment/super
COPY --from=nodebuild /workdir/public /deployment/web/public
ENV GIN_MODE=release
ENTRYPOINT ["/deployment/super"]
