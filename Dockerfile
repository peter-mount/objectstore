# ============================================================
# Dockerfile used to build the objectstore microservice
# ============================================================

ARG arch=amd64
ARG goos=linux

# ============================================================
# Build container containing our pre-pulled libraries.
# As this changes rarely it means we can use the cache between
# building each microservice.
FROM golang:alpine as build

# The golang alpine image is missing git so ensure we have additional tools
RUN apk add --no-cache \
      curl \
      git \
      tzdata

# Ensure we have the libraries - docker will cache these between builds
RUN go get -v \
      github.com/minio/minio-go/pkg/s3signer \
      github.com/peter-mount/golib/... \
      gopkg.in/mgo.v2/bson

# ============================================================
# source container contains the source as it exists within the
# repository.
FROM build as source
WORKDIR /go/src/github.com/peter-mount/objectstore
ADD . .

# ============================================================
# Compile the source.
FROM source as compiler
ARG arch
ARG goos
ARG goarch
ARG goarm

# NB: CGO_ENABLED=0 forces a static build
RUN CGO_ENABLED=0 \
    GOOS=${goos} \
    GOARCH=${goarch} \
    GOARM=${goarm} \
    go build \
      -o /dest/bin/objectstore \
      github.com/peter-mount/objectstore/bin

# ============================================================
# This is the final image
FROM alpine
RUN apk add --no-cache tzdata
COPY --from=compiler /dest/ /
