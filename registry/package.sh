#!/usr/bin/env bash

set -e

RELEASE_PATH=build/release
VERSIONS_PATH=build/providers/spacelift-io/spacelift
VERSION="1.0.0"

rm -rf build
mkdir -p $RELEASE_PATH
mkdir -p $VERSIONS_PATH

cp registry/versions.json $VERSIONS_PATH/versions

BASE_NAME=terraform-provider-spacelift
HOSTNAME=tfproviders.spacelift.io
DIGESTS_FILE=build/${BASE_NAME}_${VERSION}_SHA256SUMS

GPG_KEY_ID=242AE94ABED6D1D91DF6B3E63053164C83E7F916
GPG_ASCII_ARMOR=$(gpg --export --armor ${GPG_KEY_ID})

# Build function.
build () {
    OS=$1
    ARCH=$2

    echo "Compiling for ${OS} on ${ARCH}..." 1>&2

    BINARY_PATH=${RELEASE_PATH}/${BASE_NAME}_v${VERSION}_x4
    ZIP_NAME=${BASE_NAME}_${VERSION}_${OS}_${ARCH}.zip
    ZIP_PATH=${RELEASE_PATH}/${ZIP_NAME}

    # Step 1: build.
    CGO_ENABLED=0 \
    GOOS=$OS \
    GOARCH=$ARCH \
    go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o $BINARY_PATH

    # Step 2: zip and remove source binary
    zip $ZIP_PATH $BINARY_PATH
    rm $BINARY_PATH

    # Step 3: write SHA to the sums file.
    SHASUM=$(openssl dgst -sha256 ${ZIP_PATH} | cut -d' ' -f2)
    echo "${ZIP_NAME}  ${SHASUM}" >> $DIGESTS_FILE

    # Step 4: Add JSON manifest file.
    VERSION_DIR=${VERSIONS_PATH}/${VERSION}/download/${OS}
    mkdir -p $VERSION_DIR

    ARCH=${ARCH} \
    OS=${OS} \
    BASE_NAME=${BASE_NAME} \
    VERSION=$VERSION \
    HOSTNAME=${HOSTNAME} \
    GPG_KEY_ID=${GPG_KEY_ID} \
    SHASUM=${SHASUM} \
    GPG_ASCII_ARMOR=${GPG_ASCII_ARMOR//$'\n'/'\n'} \
    envsubst < registry/version.json > $VERSION_DIR/${ARCH}
}

# macOS
build "darwin" "amd64"

# FreeBSD
build "freebsd" "386"
build "freebsd" "amd64"
build "freebsd" "arm"

# Linux
build "linux" "386"
build "linux" "amd64"
build "linux" "arm"

# OpenBSD
build "openbsd" "386"
build "openbsd" "amd64"

# Solaris
build "solaris" "amd64"

# Windows
build "windows" "386"
build "windows" "amd64"

echo "Signing the checksums file..." 1>&2

gpg -u ${GPG_KEY_ID} \
    --sign \
    --output $DIGESTS_FILE.sig \
    --passphrase=$GPG_PASSPHRASE \
    --pinentry-mode=loopback \
    $DIGESTS_FILE
