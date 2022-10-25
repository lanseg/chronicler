load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "twitter",
    srcs = [
        "twitter.go",
        "twitter_types.go",
    ],
    importpath = "chronist/twitter",
    deps = [
        "//util",
    ],
)
