package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_binary(
    name = "main",
    srcs = ["main.go"],
    deps = [":webdriver"],
)

go_library(
    name = "webdriver",
    srcs = [
        "geckodriver.go",
        "requestbuilder.go",
        "scenario_webdriver.go",
        "scenarios.go",
        "webdriver.go",
    ],
    importpath = "chronicler/webdriver",
    deps = [
        "//util",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//concurrent",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "webdriver_test",
    size = "small",
    srcs = [
        "browser_test.go",
        "scenarios_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":webdriver",
    ],
)
