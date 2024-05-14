FROM golang:1.22 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app
VOLUME /data/conf

COPY ./configs/ /data/conf  

WORKDIR /app

EXPOSE 8000
EXPOSE 9000

CMD ["./pt", "-conf", "/data/conf"]
