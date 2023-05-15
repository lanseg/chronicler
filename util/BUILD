load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "util",
    srcs = [
        "collections.go",
        "common.go",
        "logging.go",
        "runners.go",
        "textdata.go",
    ],
    importpath = "chronicler/util",
)

go_test(
    name = "util_test",
    size = "small",
    srcs = [
        "collections_test.go",
        "runners_test.go",
        "textdata_test.go",
    ],
    embed = [
        ":util",
    ],
)
