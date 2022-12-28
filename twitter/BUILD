load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test", "go_binary")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "twitter",
    srcs = [
        "twitter.go",
    ],
    importpath = "chronist/twitter",
    deps = [
        "//util",
    ],
)

go_binary(
    name = "main",
    srcs = ["main.go"],
    deps = [
        ":twitter",
        "//util",
    ],
)
