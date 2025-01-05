FROM alpine:latest
LABEL Maintainer="lans.rf@gmail.com"

ENV STORAGESERVER_PORT=1234 \
    STATUSSERVER_PORT=1234 \
    STORAGE_ROOT=/chronicler/storage/data/

RUN apk add gcompat

WORKDIR /chronicler/storage
COPY server /chronicler/storage/server

ENTRYPOINT [ "sh", "-c", "/chronicler/storage/server --StorageServerPort=$STORAGESERVER_PORT --StorageRoot $STORAGE_ROOT"  ]
