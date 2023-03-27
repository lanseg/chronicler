load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "maybe",
    srcs = [
        "maybe.go",
    ],
    importpath = "chronicler/maybe",
)

go_test(
    name = "util_test",
    size = "small",
    srcs = [
        "maybe_test.go",
    ],
    embed = [
        ":maybe",
    ],
)
