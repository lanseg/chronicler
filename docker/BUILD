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
        ":docker_files",
    ],
    outs = ["chronicler-docker.tar"],
    cmd = """
        mkdir chronicler
        tar -xvf $(locations //:package) -C chronicler
        cp -v $(locations :docker_files) ./

        mkdir  docker-tmp
        export HOME=`pwd`/docker-tmp
        docker build --compress -t lanseg/chronicler --no-cache --output type=tar,dest=$(OUTS) .
    """,
)
