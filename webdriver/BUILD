load("@io_bazel_rules_go//go:def.bzl", "go_library")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "webdriver",
    srcs = [
        "firefox.go",
        "marionette.go",
        "webdriver.go",
    ],
    importpath = "chronicler/webdriver",
    deps = [
        "//util",
        "@commons//collections",
        "@commons//optional",
        "@commons//common",
    ],
)
