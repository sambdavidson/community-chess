FROM gosrcbase:latest as builder
ARG reposrc
COPY ./gameserver ./gameserver
RUN go get ./gameserver/...
RUN cd ./gameserver && go build .

FROM alpine:latest
ARG reposrc
COPY --from=builder /go/src/${reposrc}/gameserver/gameserver .
EXPOSE 8070 8080 8090
CMD ["./gameserver"]