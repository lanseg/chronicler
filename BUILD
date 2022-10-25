load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_source")

go_binary(
    name = "main",
    srcs = ["main.go"],
    deps = [
        "//chronist",
        "//storage",
        "//telegram",
        "//twitter",
        "//util",
    ],
)
