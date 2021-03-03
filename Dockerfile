FROM golang:1.16-alpine as builder

RUN apk add --no-cache git
ARG DIR=/project
COPY go.* $DIR/
WORKDIR $DIR
RUN go mod download
COPY . $DIR/
RUN CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /terraform-provider-spacelift

FROM alpine:3.13.2

RUN apk add --no-cache ca-certificates curl git openssh

COPY --from=builder /terraform-provider-spacelift /bin/terraform-provider-spacelift
COPY --from=builder /terraform-provider-spacelift /plugins/registry.spacelift.io/spacelift-io/spacelift/1.0.0/linux_amd64/terraform-provider-spacelift

RUN echo "hosts: files dns" > /etc/nsswitch.conf \
  && adduser --disabled-password --uid=1983 spacelift \
  && chown -R spacelift /plugins

USER spacelift
