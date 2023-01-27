load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "twitter",
    srcs = [
        "twitter.go",
    ],
    importpath = "chronicler/twitter",
    deps = [
        "//util",
    ],
)

go_test(
    name = "twitter_test",
    size = "small",
    srcs = [
        "twitter_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":twitter",
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
