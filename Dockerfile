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


RUN adduser --disabled-password --uid=1983 spacelift

ARG TF013_PROVIDER_PATH=/home/spacelift/.terraform.d/plugins/registry.spacelift.io/spacelift-io/spacelift/1.0/linux_amd64
RUN mkdir -p $TF013_PROVIDER_PATH
RUN ln -s /bin/terraform-provider-spacelift $TF013_PROVIDER_PATH/terraform-provider-spacelift

USER spacelift
