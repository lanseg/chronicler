load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "storage",
    srcs = [
        "http_downloader.go",
        "relative_fs.go",
        "storage.go",
    ],
    importpath = "chronicler/storage",
    deps = [
        "//firefox",
        "//records",
        "//records:records_go_proto",
        "//util",
        "@commons//collections",
        "@commons//optional",
    ],
)

go_test(
    name = "storage_test",
    size = "small",
    srcs = [
        "relative_fs_test.go",
        "storage_test.go",
    ],
    embed = [
        ":storage",
    ],
)
