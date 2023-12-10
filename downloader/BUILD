load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "downloader",
    srcs = [
        "downloader.go"
    ],
    importpath = "chronicler/downloader",
    deps = [
        "@commons//collections",
        "@commons//optional",
        "@commons//common",
        "@commons//concurrent",
    ],
)

go_test(
    name = "downloader_test",
    size = "small",
    srcs = [
        "downloader_test.go"
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":downloader",
    ]
)
