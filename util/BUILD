load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "util",
    importpath = "chronist/util",
    srcs = [
      "logging.go",
      "maps.go",
      "runners.go",
    ],
)

go_test(
    name = "util_test",
    srcs = [
      "maps_test.go",
    ],
    embed = [
      ":util",
    ],
)
