load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "chronicler",
    srcs = [
        "chronicler.go",
        "twitter.go",
        "web.go",
    ],
    importpath = "chronicler",
    deps = [
        "//web:htmlparser",
        "@commons//collections",
        "//proto:records_go_proto",
        "//twitter",
        "//util",
    ],
)

go_test(
    name = "chronicler_test",
    size = "small",
    srcs = [
        "chronicler_test.go",
    ],
    embed = [
        ":chronicler",
    ],
    deps = [
        "//proto:records_go_proto",
    ],
)
