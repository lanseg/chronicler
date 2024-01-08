load("@io_bazel_rules_go//go:def.bzl", "go_library")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "telegram",
    srcs = [
        "bot.go",
        "utils.go",
    ],
    importpath = "chronicler/telegram",
    deps = [
        "//util",
        "@tgbot//:telegram_bot",
        "@commons//common",
    ],
)
