load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")


go_library(
    name = "firefox",
    srcs = [
        "marionette.go",
        "firefox.go",
    ],
    importpath = "chronicler/firefox",
    deps = [
        "//util",
        "@commons//collections",
        "@commons//optional",
    ],
)

go_binary(
    name = "main",
    srcs = [
        "main.go",
    ],
    deps = [
        ":firefox",
        "//util",
        "@commons//collections",
        "@commons//optional",
    ],
)
