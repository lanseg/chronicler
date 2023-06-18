load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "server",
    srcs = [
        "server.go",
    ],
    importpath = "chronicler/frontend",
    deps = [
        "//proto:records_go_proto",
        "//storage",
        "//util",
        "//web:htmlparser",
        "@commons//collections",
    ],
)

go_binary(
    name = "frontend",
    srcs = ["main.go"],
    data = glob(["static/*"]),
    deps = [
        ":server",
        "//proto:records_go_proto",
        "//util",
    ],
)

go_test(
    name = "frontend_test",
    size = "small",
    srcs = [
        "server_test.go",
    ],
    embed = [
        ":server",
    ],
)
