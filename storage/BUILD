load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "storage",
    importpath = "chronist/storage",
    deps = [
      "//util:util",
    ],
    srcs = [
      "storage.go",
    ],
)
