load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "webdriver",
    srcs = [
        "firefox.go",
        "marionette.go",
        "scenarios.go",
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

go_test(
    name = "webdriver_test",
    size = "small",
    srcs = [
        "scenarios_test.go"
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":webdriver",
    ]
)
