FROM alpine:latest
WORKDIR chronicler
COPY chronicler/frontend .
COPY chronicler/frontendserver .
COPY chronicler/main .
COPY chronicler/scenarios .
COPY chronicler/statusserver .
COPY chronicler/storageserver .
COPY entrypoint.sh .
RUN ["apk", "add", "libc6-compat", "bash"]
EXPOSE 1235
ENTRYPOINT ["bash", "entrypoint.sh"]