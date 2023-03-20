load("@io_bazel_rules_go//go:def.bzl", "go_library")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "storage",
    srcs = [
        "storage.go",
    ],
    importpath = "chronicler/storage",
    deps = [
        "//proto:records_go_proto",
        "//util",
    ],
)
