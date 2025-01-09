package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "adapter",
    srcs = [
        "adapter.go",
    ],
    importpath = "chronicler/adapter",
    deps = [
        "//records:records_go_proto",
    ],
)

go_test(
    name = "adapter_test",
    size = "small",
    srcs = [
        "adapter_test.go",
    ],
    embed = [
        ":adapter",
    ],
    deps = [
        "//records:records_go_proto",
    ],
)
