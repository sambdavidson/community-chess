FROM gosrcbase:latest as builder
ARG REPOSRC
COPY ./debugwebserver ./debugwebserver
RUN go get ./debugwebserver/...
RUN cd ./debugwebserver && go build .

FROM alpine:latest
ARG REPOSRC
COPY --from=builder /go/src/${REPOSRC}/debugwebserver/debugwebserver .
COPY --from=builder /go/src/${REPOSRC}/debugwebserver/static ./static
EXPOSE 80
CMD ["./debugwebserver", "--help"]