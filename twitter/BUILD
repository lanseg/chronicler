load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "twitter",
    importpath = "chronist/twitter",
    srcs = [
      "twitter.go",
      "twitter_types.go",
    ],
    deps = [
      "//util:util",
    ],
)
 
