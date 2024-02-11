package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "storage",
    srcs = [
        "overlay.go",
        "storage.go",
    ],
    importpath = "chronicler/storage",
    deps = [
        "//downloader",
        "//records",
        "//records:records_go_proto",
        "//util",
        "//webdriver",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "storage_test",
    size = "small",
    srcs = [
        "overlay_test.go",
        "storage_test.go",
    ],
    embed = [
        ":storage",
    ],
)
