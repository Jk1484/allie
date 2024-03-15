FROM golang:1.21.5-alpine as build

RUN mkdir /allie

ADD . /allie

WORKDIR /allie

RUN go build -o allie ./cmd

FROM alpine:latest
COPY --from=build /allie /allie

WORKDIR /allie

CMD ["/allie/allie"]