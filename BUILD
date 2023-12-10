load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = [
        "main.go",
    ],
    deps = [
        "//adapter",
        "//downloader",
        "//records:records_go_proto",
        "//storage",
        "//telegram",
        "//twitter",
        "//util",
        "//webdriver",
        "@commons//collections",
        "@commons//common",
        "@commons//optional",
    ],
)
