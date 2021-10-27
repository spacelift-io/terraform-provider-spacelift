FROM golang:1.16-alpine as builder

ARG VERSION=dev
ARG COMMIT=dev

# 3rd party soft dependency versions
ARG INFRACOST_VERSION=0.9.7
ARG TERRAGRUNT_VERSION=0.35.4

RUN apk add --no-cache curl git

# Download infracost
RUN curl -s -L https://github.com/infracost/infracost/releases/download/v${INFRACOST_VERSION}/infracost-linux-amd64.tar.gz | \
  tar xz -C /tmp && \
  mv /tmp/infracost-linux-amd64 /bin/infracost

# Download Terragrunt.
RUN wget -O /bin/terragrunt https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64 \
    && chmod +x /bin/terragrunt

ARG DIR=/project
COPY go.* $DIR/
WORKDIR $DIR
RUN go mod download
COPY . $DIR/

RUN CGO_ENABLED=0 go build -a -tags netgo -ldflags "-w -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /terraform-provider-spacelift

FROM alpine:3.14.0

RUN apk -U upgrade
RUN apk add --no-cache ca-certificates curl git openssh jq aws-cli

COPY --from=builder /bin/infracost /bin/infracost
COPY --from=builder /bin/terragrunt /bin/terragrunt
COPY --from=builder /terraform-provider-spacelift /bin/terraform-provider-spacelift
COPY --from=builder /terraform-provider-spacelift /plugins/registry.spacelift.io/spacelift-io/spacelift/1.0.0/linux_amd64/terraform-provider-spacelift

RUN echo "hosts: files dns" > /etc/nsswitch.conf \
  && adduser --disabled-password --uid=1983 spacelift \
  && chown -R spacelift /plugins

USER spacelift
