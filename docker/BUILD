package(default_visibility = ["//visibility:public"])

filegroup(
    name="docker_files",
    srcs = [
        "Dockerfile",
        "entrypoint.sh"
    ],
)


genrule(
    name = "docker",
    srcs = [
        "//:package",
        "//third_party/telegram:telegram-bot-api",
        ":docker_files",
    ],
    outs = ["chronicler-docker.tar"],
    cmd = """
        mkdir chronicler
        tar -xvf $(locations //:package) -C chronicler
        cp -v $(locations :docker_files) ./
        cp -v $(locations //third_party/telegram:telegram-bot-api) ./

        mkdir  ttmp
        export HOME=`pwd`/docker-tmp
        docker build --compress -t lanseg/chronicler --no-cache --output type=tar,dest=$(OUTS) .
    """,
)
