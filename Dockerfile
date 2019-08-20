FROM golang:1.12-alpine as builder

COPY . /project
WORKDIR /project
RUN apk add --no-cache git \
  && go build -o /terraform-provider-spacelift

FROM alpine:3.10

RUN apk add --no-cache ca-certificates curl git openssh
COPY --from=builder /terraform-provider-spacelift /usr/bin/terraform-provider-spacelift

USER nobody
