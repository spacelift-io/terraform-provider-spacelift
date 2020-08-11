FROM golang:1.13-alpine as builder

RUN apk add --no-cache git
ARG DIR=/project
COPY go.* $DIR/
WORKDIR $DIR
RUN go mod download
COPY . $DIR/
RUN CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /terraform-provider-spacelift

FROM alpine:3.10

RUN apk add --no-cache ca-certificates curl git openssh

COPY --from=builder /terraform-provider-spacelift /bin/terraform-provider-spacelift

RUN echo "hosts: files dns" > /etc/nsswitch.conf

RUN adduser --disabled-password --no-create-home --uid=1983 spacelift

USER spacelift
