load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "tinyhtml",
    srcs = [
        "tokenizer.go",
        "html.go"
    ],
    importpath = "github.com/lanseg/golang-commons/tinyhtml",
    deps = [
        "@commons//optional",
        "@commons//collections",
    ],
)

go_test(
    name = "tinyhtml_test",
    size = "small",
    srcs = [
        "tinyhtml_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":tinyhtml",
    ],
)
