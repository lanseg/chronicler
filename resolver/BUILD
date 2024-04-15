load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "resolver",
    srcs = [
        "resolver.go",
        "downloader.go"
    ],
    importpath = "chronicler/resolver",
    visibility = ["//visibility:public"],
    deps = [
        "//records:records_go_proto",
        "//storage",
        "//webdriver",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//concurrent",
        "@golang-commons//optional",
    ],
)
go_test(
    name = "resolver_test",
    size = "small",
    srcs = [
        "downloader_test.go",
        "resolver_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":resolver",
    ],
)
