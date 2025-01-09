FROM alpine:latest
LABEL Maintainer="lans.rf@gmail.com"

ENV WEBDRIVER_PORT=1235 \
    ALLOWED_HOSTS=resolver

RUN apk add firefox geckodriver

EXPOSE $WEBDRIVER_PORT
ENTRYPOINT [ "sh", "-c", "geckodriver -p $WEBDRIVER_PORT --host webdriver --allow-hosts=$ALLOWED_HOSTS -vvv " ]
