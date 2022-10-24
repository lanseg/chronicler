load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "chronist",
    srcs = ["chronist.go"],
    importpath = "chronist",
    deps = [
      "//util:util",
      "//storage:storage",
      "//telegram:telegram",
      "//twitter:twitter",
      "//proto:records_go_proto",
    ],
)

go_test(
    name = "chronist_test",
    srcs = [
      "chronist_test.go",
    ],
    embed = [
      ":chronist",
    ],
)
