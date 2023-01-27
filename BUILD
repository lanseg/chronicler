load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = ["main.go"],
    deps = [
        "//chronicler",
        "//proto:records_go_proto",
        "//storage",
        "//telegram",
        "//twitter",
        "//util",
    ],
)
