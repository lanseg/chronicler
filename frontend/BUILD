load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "server",
    srcs = [
        "server.go",
    ],
    importpath = "chronicler/frontend",
    deps = [
        "//webdriver",
        "//records:records_go_proto",
        "//storage",
        "//util",
        "//web:htmlparser",
        "@commons//collections",
        "@commons//common",
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
        "@commons//common",
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
    ],
    embed = [
        ":server",
    ],
)
