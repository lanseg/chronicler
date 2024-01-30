package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "util",
    srcs = [
        "runners.go",
        "textdata.go",
    ],
    importpath = "chronicler/util",
    deps = [
      "@golang-commons//common",
      "@golang-commons//optional",
    ],
)

go_test(
    name = "util_test",
    size = "small",
    srcs = [
        "runners_test.go",
        "textdata_test.go",
    ],
    embed = [
        ":util",
    ],
)
