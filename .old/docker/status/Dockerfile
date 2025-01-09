FROM alpine:latest
LABEL Maintainer="lans.rf@gmail.com"

ENV STATUSSERVER_PORT=1234
RUN apk add gcompat

WORKDIR /chronicler/status
COPY server /chronicler/status/server

ENTRYPOINT [ "sh", "-c", "/chronicler/status/server --StatusServerPort $STATUSSERVER_PORT"  ]
