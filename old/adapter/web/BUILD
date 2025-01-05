package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "web",
    srcs = [
        "web.go",
    ],
    importpath = "chronicler/adapter/web",
    deps = [
        "//adapter",
        "//records",
        "//records:records_go_proto",
        "//webdriver",
        "@golang-commons//almosthtml",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "web_test",
    size = "small",
    srcs = [
        "web_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":web",
    ],
    deps = [
        "//records:records_go_proto",
    ],
)
