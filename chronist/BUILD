load("@io_bazel_rules_go//go:def.bzl", "go_library")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "chronist",
    srcs = ["chronist.go"],
    importpath = "chronist",
    deps = [
      "//util:util",
      "//storage:storage",
      "//telegram:telegram",
      "//proto:records_go_proto",
    ],
)
