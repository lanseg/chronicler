package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "twitter",
    srcs = [
        "twitter.go",
    ],
    importpath = "chronicler/twitter",
    deps = [
        "//util",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "twitter_test",
    size = "small",
    srcs = [
        "twitter_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":twitter",
    ],
)
