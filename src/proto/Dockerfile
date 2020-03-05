FROM znly/protoc:0.4.0
ARG REPOSRC
WORKDIR ${REPOSRC}
COPY . ./proto
RUN find ~ -type f -name "*.proto" -exec protoc  --proto_path=/ --go_out=plugins=grpc:. {} +
ENTRYPOINT ["top", "-b"]