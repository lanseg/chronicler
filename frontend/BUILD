package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test", "go_binary")

go_library(
    name = "server",
    srcs = [
        "server.go",
        "pathparam_handler.go",
    ],
    importpath = "chronicler/frontend",
    deps = [
        "//downloader",
        "//webdriver",
        "//records",
        "//records:records_go_proto",
        "//storage",
        "//util",
        "@golang-commons//almosthtml",
        "@golang-commons//collections",
        "@golang-commons//common",
    ],
)

go_binary(
    name = "frontend",
    srcs = ["main.go"],
    data = glob([
        "static/*",
        "static/modules/*",
    ]),
    deps = [
        ":server",
        "//records:records_go_proto",
        "//util",
        "@golang-commons//common",
    ],
)

go_test(
    name = "frontend_test",
    size = "small",
    data = glob([
        "static/*",
        "static/modules/*",
    ]),
    srcs = [
        "server_test.go",
        "pathparam_handler_test.go",
    ],
    embed = [
        ":server",
    ],
)
