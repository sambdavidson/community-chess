FROM gosrcbase:latest as builder
ARG REPOSRC
COPY ./gameserver ./gameserver
RUN go get ./gameserver/...
RUN cd ./gameserver && go build .

FROM alpine:latest
ARG REPOSRC
COPY --from=builder /go/src/${REPOSRC}/gameserver/gameserver .
EXPOSE 8070 8080 8090
CMD ["./gameserver", "--help"]