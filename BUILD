load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = [
        "config.go",
        "main.go",
    ],
    deps = [
        "//adapter",
        "//proto:records_go_proto",
        "//storage",
        "//telegram",
        "//twitter",
        "//util",
        "//web:htmlparser",
        "@commons//collections",
    ],
)
