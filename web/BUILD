load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "htmlparser",
    srcs = [
        "htmlparser.go",
    ],
    importpath = "web/htmlparser",
    deps = [
        ":tokenizer",
        "//util",
        "@commons//collections",
    ],
)

go_library(
    name = "tokenizer",
    srcs = [
        "tokenizer.go",
    ],
    importpath = "web/tokenizer",
    deps = [
        "//util",
    ],
)

go_test(
    name = "web_test",
    size = "small",
    srcs = [
        "tokenizer_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":tokenizer",
    ],
)
