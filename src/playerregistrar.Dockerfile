FROM gosrcbase:latest as builder
ARG reposrc
COPY ./playerregistrar ./playerregistrar
RUN go get ./playerregistrar/...
RUN cd ./playerregistrar && go build .

FROM alpine:latest
ARG reposrc
COPY --from=builder /go/src/${reposrc}/playerregistrar/playerregistrar .
EXPOSE 8080
CMD ["./playerregistrar"]