load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "chronicler",
    srcs = [
        "chronicler.go",
        "config.go",
        "telegram.go",
        "twitter.go",
        "web.go",
    ],
    importpath = "chronicler",
    deps = [
        "//proto:records_go_proto",
        "//telegram",
        "//twitter",
        "//util",
        "//web:htmlparser",
        "@commons//collections",
        "@commons//optional",
    ],
)

go_test(
    name = "chronicler_test",
    size = "small",
    srcs = [
        "chronicler_test.go",
        "telegram_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":chronicler",
    ],
    deps = [
        "//proto:records_go_proto",
    ],
)
