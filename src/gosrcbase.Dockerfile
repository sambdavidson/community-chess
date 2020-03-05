FROM proto:latest as protos

FROM golang:alpine
ARG REPOSRC
WORKDIR /go/src${REPOSRC}
COPY --from=protos ${REPOSRC}/proto ./proto
COPY ./lib ./lib
RUN apk add --no-cache git
RUN go get google.golang.org/grpc
ENTRYPOINT ["top", "-b"]