package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "downloader",
    srcs = [
        "downloader.go",
    ],
    importpath = "chronicler/downloader",
    deps = [
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//concurrent",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "downloader_test",
    size = "small",
    srcs = [
        "downloader_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":downloader",
    ],
)
