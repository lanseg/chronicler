package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "webdriver",
    srcs = [
        "firefox.go",
        "marionette.go",
        "browser.go",
        "scenarios.go",
        "scenario_webdriver.go",
        "webdriver.go",
    ],
    importpath = "chronicler/webdriver",
    deps = [
        "//util",
        "@golang-commons//collections",
        "@golang-commons//optional",
        "@golang-commons//common",
        "@golang-commons//concurrent",
    ],
)

go_test(
    name = "webdriver_test",
    size = "small",
    srcs = [
        "scenarios_test.go",
        "browser_test.go"
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":webdriver",
    ]
)
