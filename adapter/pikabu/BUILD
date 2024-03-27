package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "pikabu",
    srcs = [
        "pikabu.go",
        "parser.go",
        "sourceprovider.go",
    ],
    importpath = "chronicler/adapter/pikabu",
    deps = [
        "//adapter",
        "//records",
        "//records:records_go_proto",
        "//webdriver",
        "@org_golang_x_text//encoding/charmap",
        "@golang-commons//almosthtml",
        "@golang-commons//collections",
        "@golang-commons//common",
    ],
)

go_test(
    name = "pikabu_test",
    size = "small",
    srcs = [
        "parser_test.go",
        "sourceprovider_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [
        ":pikabu",
    ],
    deps = [
        "//records:records_go_proto",
    ],
)
