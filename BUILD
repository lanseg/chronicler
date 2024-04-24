load("@rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = [
        "main.go",
    ],
    deps = [
        "//adapter",
        "//adapter/pikabu",
        "//adapter/telegram",
        "//adapter/twitter",
        "//adapter/web",
        "//chronicler",
        "//resolver",
        "//records:records_go_proto",
        "//status",
        "//status:status_go_proto",
        "//storage",
        "//storage/endpoint",
        "//storage/endpoint:storage_endpoint_go_proto",
        "//util",
        "//webdriver",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//concurrent",
        "@golang-commons//optional",
        "@tgbot//:telegram_bot",
    ],
)

genrule(
    name = "package",
    srcs = [
        ":main",
        "//scenarios:scenarios",
        "//frontend:frontendserver",
        "//frontend:static_files",
        "//storage/endpoint:storageserver",
        "//status:statusserver"
    ],
    outs = ["chronicler.tar.gz"],
    cmd = """
        tar -chzvf $@ \
          --transform 's/bazel-out.*_\\///g' \
          $(location :main) \
          $(locations //scenarios:scenarios) \
          $(location //frontend:frontendserver) \
          $(location //storage/endpoint:storageserver) \
          $(location //status:statusserver) \
          $(locations //frontend:static_files)
    """,
)