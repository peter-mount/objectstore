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
      tzdata \
      zip

# Ensure we have the libraries - docker will cache these between builds
#RUN go get -v \
#      github.com/peter-mount/go-glob \
#      github.com/peter-mount/golib/... \
#      github.com/peter-mount/sortfold \
#      gopkg.in/mgo.v2/bson \
#      gopkg.in/robfig/cron.v2 \
#      gopkg.in/yaml.v2

# ============================================================
# source container contains the source as it exists within the
# repository.
FROM build as source
#WORKDIR /go/src/github.com/peter-mount/objectstore
WORKDIR /work

# Download dependencies before copying any sources then we
# can use the docker cache to limit updates
ADD go.mod .
RUN go mod download

ADD . .

# ============================================================
# Run all tests in a new container so any output won't affect
# the final build.
FROM source as test

WORKDIR /work
RUN go test -v \
      github.com/peter-mount/objectstore/policy \
      github.com/peter-mount/objectstore/utils

# ============================================================
# Compile the source.
FROM source as compiler
ARG arch
ARG goos
ARG goarch
ARG goarm
WORKDIR /work

# NB: CGO_ENABLED=0 forces a static build
RUN CGO_ENABLED=0 \
    GOOS=${goos} \
    GOARCH=${goarch} \
    GOARM=${goarm} \
    go build \
      -o /dest/bin/objectstore \
      github.com/peter-mount/objectstore/bin

# ============================================================
# Optional stage, upload the binaries as a tar file
FROM compiler AS upload
ARG uploadPath=
ARG uploadCred=
ARG uploadName=
RUN if [ -n "${uploadCred}" -a -n "${uploadPath}" -a -n "${uploadName}" ] ;\
    then \
      cd /dest/bin; \
      tar cvzpf /tmp/${uploadName}.tgz * && \
      zip /tmp/${uploadName}.zip * && \
      curl -u ${uploadCred} --upload-file /tmp/${uploadName}.tgz ${uploadPath}/ && \
      curl -u ${uploadCred} --upload-file /tmp/${uploadName}.zip ${uploadPath}/; \
    fi

# ============================================================
# This is the final image
FROM alpine
RUN apk add --no-cache tzdata
COPY --from=compiler /dest/ /
