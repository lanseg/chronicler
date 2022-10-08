load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "telegram",
    importpath = "chronist/telegram",
    deps = [
      "//util:util",
    ],
    srcs = [
      "telegram.go",
      "telegram_types.go",
    ],
)
