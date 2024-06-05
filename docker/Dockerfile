FROM archlinux:latest
WORKDIR chronicler
COPY chronicler/telegram-bot-api .
COPY chronicler/frontend .
COPY chronicler/frontendserver .
COPY chronicler/main .
COPY chronicler/scenarios .
COPY chronicler/statusserver .
COPY chronicler/storageserver .
COPY entrypoint.sh .
COPY telegram .
EXPOSE 1235
ENTRYPOINT ["bash", "entrypoint.sh"]
