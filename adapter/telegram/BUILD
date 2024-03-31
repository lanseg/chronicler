package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "telegram",
    srcs = [
        "telegram.go",
    ],
    importpath = "chronicler/adapter/telegram",
    deps = [
        "//adapter",
        "//records",
        "//records:records_go_proto",
        "//util",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
        "@tgbot//:telegram_bot",
    ],
)

go_test(
    name = "telegram_test",
    size = "small",
    srcs = [
        "telegram_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [":telegram"],
    deps = ["//records:records_go_proto"],
)
