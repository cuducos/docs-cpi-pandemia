FROM golang:1.16-buster as build
WORKDIR /docs-cpi-pandemia
ADD go.* ./
ADD main.go .
ADD cache/ ./cache/
ADD cli/ ./cli/
ADD downloader/ ./downloader/
ADD filesystem/ ./filesystem/
RUN go get && go build -o /usr/bin/docs-cpi-pandemia

FROM debian:buster-slim
COPY --from=build /usr/bin/docs-cpi-pandemia /usr/bin/docs-cpi-pandemia
CMD ["docs-cpi-pandemia"]
