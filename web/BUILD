load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_binary(
    name = "main",
    srcs = ["main.go"],
    deps = [
        ":tokenizer",
    ],
)

go_library(
    name = "tokenizer",
    srcs = [
        "tokenizer.go",
    ],
    importpath = "web/tokenizer",
    deps = [],
)
