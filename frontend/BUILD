package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "server",
    srcs = [
        "pathparam_handler.go",
        "server.go",
    ],
    importpath = "chronicler/frontend",
    deps = [
        "//records",
        "//records:records_go_proto",
        "//status",
        "//storage",
        "//storage/endpoint",
        "//storage/endpoint:storage_endpoint_go_proto",
        "//util",
        "//webdriver",
        "@golang-commons//almosthtml",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
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
        "//storage/endpoint",
        "//status",
        "//util",
        "@golang-commons//common",
    ],
)

go_test(
    name = "frontend_test",
    size = "small",
    srcs = [
        "pathparam_handler_test.go",
        "server_test.go",
    ],
    data = glob([
        "static/*",
        "static/modules/*",
    ]),
    embed = [
        ":server",
    ],
)
