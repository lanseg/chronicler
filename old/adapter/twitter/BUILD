package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "twitter",
    srcs = [
        "client.go",
        "twitter.go",
    ],
    importpath = "chronicler/adapter/twitter",
    deps = [
        "//adapter",
        "//records:records_go_proto",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "twitter_test",
    size = "small",
    srcs = [
        "client_test.go",
        "twitter_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":twitter",
    ],
    deps = [
        "//records:records_go_proto",
    ],
)
