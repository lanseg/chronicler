package(default_visibility = ["//visibility:public"])

genrule(
    name = "telegram-bot-api",
    srcs = [],
    outs = ["telegram-bot-api"],
    cmd = """
        mkdir {build,install}
        git clone --recursive https://github.com/tdlib/telegram-bot-api.git
        export CCACHE_DIR=/tmp/ccache
        
        cd build
        cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=../install ../telegram-bot-api/
        cmake --build . --target install -j`nproc`
        cd ..
        
        cp -rfv "install/bin/telegram-bot-api" "$(OUTS)"
    """,
)
