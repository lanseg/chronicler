load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "adapter",
    srcs = [
        "adapter.go",
        "telegram.go",
        "twitter.go",
        "web.go",
    ],
    importpath = "chronicler/adapter",
    deps = [
        "//firefox",
        "//records:records_go_proto",
        "//storage",
        "//telegram",
        "//twitter",
        "//util",
        "//web:htmlparser",
        "@commons//collections",
        "@commons//optional",
    ],
)

go_test(
    name = "adapter_test",
    size = "small",
    srcs = [
        "adapter_test.go",
        "telegram_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":adapter",
    ],
    deps = [
        "//records:records_go_proto",
    ],
)
