FROM gosrcbase:latest as builder
ARG REPOSRC
COPY ./playerregistrar ./playerregistrar
RUN go get ./playerregistrar/...
RUN cd ./playerregistrar && go build .

FROM alpine:latest
ARG REPOSRC
COPY --from=builder /go/src/${REPOSRC}/playerregistrar/playerregistrar .
EXPOSE 8081
CMD ["./playerregistrar", "--help"]