package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "tiktok",
    srcs = [
        "tiktok.go",
    ],
    importpath = "chronicler/adapter/tiktok",
    deps = [
        "//adapter",
        "//records:records_go_proto",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "tiktok_test",
    size = "small",
    srcs = [
        "tiktok_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":tiktok",
    ],
    deps = [
        "//records:records_go_proto",
    ],
)
