FROM alpine:latest
LABEL Maintainer="lans.rf@gmail.com"

ENV FRONTEND_PORT=8080 \
    STATUSSERVER_PORT=1234 \
    STORAGESERVER_PORT=1234
RUN apk add gcompat

WORKDIR /chronicler/frontend
COPY static /chronicler/frontend/static
COPY server /chronicler/frontend/server
COPY entrypoint.sh /chronicler/frontend/

ENTRYPOINT [ "sh", "-x", "/chronicler/frontend/entrypoint.sh" ]
