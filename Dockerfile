FROM harbor.tianxiang.love:30443/library/alpine:latest

WORKDIR /app

RUN mkdir -pv /app/config /app/template
RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ENV TZ=Asia/Shanghai

COPY bin/prometheus-webhook-linux-amd64 /app/prometheus-webhook-linux-amd64
COPY config/config.yaml.example /app/config/config.yaml
COPY templates/ /app/templates/

EXPOSE 8080

CMD ["/app/prometheus-webhook-linux-amd64"]