FROM alpine:latest

COPY ./bugtracker /app/bin/
COPY ./config.yaml /etc/bugtracker/
COPY ./scripts/ /app/scripts/
WORKDIR /app
EXPOSE 80
RUN mkdir -p /app/upload && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && apk add tzdata && apk add bash
RUN /bin/cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo 'Asia/Shanghai' >/etc/timezone

CMD ["/bin/sh", "/app/scripts/start.sh"]
