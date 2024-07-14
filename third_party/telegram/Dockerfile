FROM alpine:latest
LABEL Maintainer="lans.rf@gmail.com"

RUN ["apk", "add", "gcompat"]

WORKDIR telegram-bot-api
COPY entrypoint.sh .
COPY telegram-bot-api .
EXPOSE 8081

ENTRYPOINT ["sh", "entrypoint.sh"]
