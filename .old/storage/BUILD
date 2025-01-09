package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "storage",
    srcs = [
        "localstorage.go",
        "storage.go",
    ],
    importpath = "chronicler/storage",
    deps = [
        "//records",
        "//records:records_go_proto",
        "//util",
        "@golang-commons//almostio",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "storage_test",
    size = "small",
    srcs = [
        "localstorage_test.go",
    ],
    embed = [
        ":storage",
    ],
)
