load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

exports_files(["update_types.py"])

genrule(
    name = "telegram_types",
    outs = ["telegram_types.go"],
    tools = ["update_types.py"],
    cmd = "python $(location update_types.py) > \"$@\"",
    visibility = ["//visibility:private"]
)

go_library(
    name = "telegram",
    importpath = "chronist/telegram",
    srcs = [
      "telegram.go",
      ":telegram_types",
    ],
    deps = [
      "//util:util",
    ],
)
