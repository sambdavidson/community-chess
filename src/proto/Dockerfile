FROM znly/protoc:0.4.0
ARG reposrc=/github.com/sambdavidson/community-chess/src
WORKDIR ${reposrc}
COPY . ./proto
RUN find ~ -type f -name "*.proto" -exec protoc  --proto_path=/ --go_out=plugins=grpc:. {} +