package(default_visibility = ["//visibility:public"])

genrule(
    name = "telegram-bot-api-client",
    srcs = [],
    outs = ["telegram-bot-api"],
    cmd = """
        mkdir {build,install}
        git clone --recursive https://github.com/tdlib/telegram-bot-api.git
        mkdir -p /tmp/chronicler_ccache
        export CCACHE_DIR=/tmp/chronicler_ccache
        
        cd build
        cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_COMPILER_LAUNCHER=ccache -DCMAKE_INSTALL_PREFIX=../install ../telegram-bot-api/
        cmake --build . --target install -j`nproc`
        cd ..
        
        cp -rfv "install/bin/telegram-bot-api" "$(OUTS)"
    """,
)

genrule(
    name = "docker",
    srcs = [
        ":telegram-bot-api-client",
        "Dockerfile",
        "entrypoint.sh",
    ],
    outs = ["telegram-bot-api_docker.tar"],
    cmd = """
        cp -v $(locations :telegram-bot-api-client) ./
        cp -v $(location Dockerfile) ./
        cp -v $(location entrypoint.sh) ./

        mkdir  docker-tmp
        export HOME=`pwd`/docker-tmp
        docker build --compress -t lanseg/telegram-bot-api --no-cache --output type=tar,dest=$(OUTS) .
    """,
)
