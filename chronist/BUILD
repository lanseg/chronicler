load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "chronist",
    srcs = [
        "chronist.go",
    ],
    deps = [
        "//proto:records_go_proto",
    ],
    importpath = "chronist",
)

go_test(
    name = "chronist_test",
    size = "small",
    srcs = [
        "chronist_test.go",
    ],
    deps = [
        "//proto:records_go_proto",
    ],
    embed = [
        ":chronist",
    ],
)
