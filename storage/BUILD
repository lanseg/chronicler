package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "storage",
    srcs = [
        "overlay.go",
        "storage.go",
        "localstorage.go",
    ],
    importpath = "chronicler/storage",
    deps = [
        "//records",
        "//records:records_go_proto",
        "//util",
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
        "localstorage_test.go",
    ],
    embed = [
        ":storage",
    ],
)
