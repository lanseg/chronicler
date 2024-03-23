load("@rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = [
        "main.go",
        "resolver.go",
    ],
    deps = [
        "//adapter",
        "//adapter/pikabu",
        "//adapter/telegram",
        "//adapter/twitter",
        "//adapter/web",
        "//downloader",
        "//records:records_go_proto",
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
