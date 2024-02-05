package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "adapter",
    srcs = [
        "adapter.go",
        "pikabu.go",
        "telegram.go",
        "twitter.go",
        "web.go",
    ],
    importpath = "chronicler/adapter",
    deps = [
        "//records",
        "//records:records_go_proto",
        "//storage",
        "//twitter",
        "//util",
        "//webdriver",
        "@golang-commons//almosthtml",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
        "@tgbot//:telegram_bot",
    ],
)

go_test(
    name = "adapter_test",
    size = "small",
    srcs = [
        "adapter_test.go",
        "pikabu_test.go",
        "telegram_test.go",
        "testutil.go",
        "twitter_test.go",
        "web_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":adapter",
    ],
    deps = [
        "//records:records_go_proto",
    ],
)
