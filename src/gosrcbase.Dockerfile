FROM proto:latest as protos

FROM golang:alpine
ARG reposrc
WORKDIR /go/src${reposrc}
COPY --from=protos ${reposrc}/proto ./proto
COPY ./lib ./lib
RUN apk add --no-cache git