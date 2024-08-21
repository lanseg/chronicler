FROM alpine:latest
LABEL Maintainer="lans.rf@gmail.com"

ENV STORAGESERVER=storageserver \
    STATUSSERVER=statusserver \
    STORAGE_ROOT=/chronicler/storage/data/ \
    TWITTER_API_KEY=some_twitter_api_key \
    TELEGRAM_API_URL=http://telegram:8081 \
    TELEGRAM_BOT_KEY=some_telegram_bot_key \
    TELEGRAM_FILE_PREFIX=some_file_prefix \
    WEBDRIVER_SERVER=some_webdriver_server \
    SCENARIO_LIBRARY=/chronicler/resolver/senarios/scenarios.json

RUN apk add gcompat

WORKDIR /chronicler/resolver
COPY resolver /chronicler/resolver/resolver
COPY scenarios /chronicler/resolver/scenarios

ENTRYPOINT [ "sh", "-c", "/chronicler/resolver/resolver --StorageServer=$STORAGESERVER --StatusServer $STATUSSERVER --TwitterApiKey $TWITTER_API_KEY --TelegramApiUrl $TELEGRAM_API_URL --TelegramBotKey $TELEGRAM_BOT_KEY --TelegramFilePrefix $TELEGRAM_FILE_PREFIX --ScenarioLibrary $SCENARIO_LIBRARY --StorageRoot $STORAGE_ROOT --WebdriverServer $WEBDRIVER_SERVER "  ]
