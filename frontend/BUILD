load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "frontend",
    srcs = [
        "server.go",
    ],
    importpath = "chronicler/frontend",
    deps = [
        "//proto:records_go_proto",
        "//util",
    ],
)

go_binary(
    name = "main",
    srcs = ["main.go"],
    data = glob(["static/*"]),
    deps = [
        ":frontend",
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
        ":frontend",
    ],
)
