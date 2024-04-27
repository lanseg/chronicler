package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "chronicler",
    srcs = [
        "chronicler.go",
        "scheduler.go",
    ],
    importpath = "chronicler/chronicler",
    deps = [
        "//adapter",
        "//resolver",
        "//records:records_go_proto",
        "//status",
        "//storage",
        "//storage/endpoint",
        "//storage/endpoint:storage_endpoint_go_proto",
        "@golang-commons//common",
        "@golang-commons//concurrent",
    ],
)

go_test(
    name = "chronicler_test",
    size = "small",
    srcs = [
        "chronicler_test.go",
        "scheduler_test.go",
    ],
    embed = [
        ":chronicler",
    ],
)
