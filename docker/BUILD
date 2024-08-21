package(default_visibility = ["//visibility:public"])

filegroup(
    name = "docker_files",
    srcs = [
        "default.env",
        "docker-compose.yaml",
        "frontend",
        "resolver",
        "status",
        "storage",
        "webdriver",
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
        set -x
        mkdir chronicler chronicler-docker

        tar -chzvf dockerfiles.tar \
            --transform 's/bazel-out.*_\\///g' \
            $(locations :docker_files) \


        tar -xvf dockerfiles.tar --strip-components=1 -C ./chronicler-docker/
        tar -xvf $(locations //:package) -C ./chronicler/
        cp -v chronicler/main chronicler-docker/resolver/resolver
        cp -rv chronicler/scenarios chronicler-docker/resolver/scenarios   
        cp -v chronicler/frontendserver chronicler-docker/frontend/server
        cp -rv chronicler/frontend/static chronicler-docker/frontend/static
        cp -v chronicler/storageserver chronicler-docker/storage/server
        cp -v chronicler/statusserver chronicler-docker/status/server

        tar -czvf $(OUTS) chronicler-docker 
    """,
)
