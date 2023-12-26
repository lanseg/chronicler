load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "status",
    srcs = [
        "status.go",
    ],
    importpath = "chronicler/status",
    deps = [
      "@commons//common",
      "@commons//optional",
    ],
)

go_test(
    name = "status_test",
    size = "small",
    srcs = [
        "status_test.go",
    ],
    embed = [
        ":status",
    ],
)
