load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "storage",
    srcs = [
        "storage.go",
    ],
    importpath = "chronist/storage",
    deps = [
        "//proto:records_go_proto",
        "//util",
        "@com_github_golang_protobuf//proto:go_default_library",
    ],
)
