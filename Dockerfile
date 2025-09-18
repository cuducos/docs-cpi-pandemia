FROM golang:1.25-trixie AS build
WORKDIR /docs-cpi-pandemia
COPY go.* ./
COPY main.go .
COPY bar/ ./bar/
COPY cache/ ./cache/
COPY cli/ ./cli/
COPY downloader/ ./downloader/
COPY filesystem/ ./filesystem/
COPY text/ ./text/
COPY unzip/ ./unzip/
RUN go get && go build -o /usr/bin/docs-cpi-pandemia

FROM debian:trixie-slim
RUN apt-get update && apt-get install ca-certificates -y
COPY --from=build /usr/bin/docs-cpi-pandemia /usr/bin/docs-cpi-pandemia
CMD ["docs-cpi-pandemia"]
