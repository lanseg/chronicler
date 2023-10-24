load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = [
        "config.go",
        "main.go",
    ],
    deps = [
        "//adapter",
        "//records:records_go_proto",
        "//storage",
        "//telegram",
        "//twitter",
        "//util",
        "//web:htmlparser",
        "//webdriver",
        "@commons//collections",
        "@commons//common",
    ],
)
