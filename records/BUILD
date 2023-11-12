load("@rules_proto//proto:defs.bzl", "proto_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

proto_library(
    name = "records_proto",
    srcs = ["records.proto"],
)

go_proto_library(
    name = "records_go_proto",
    importpath = "chronicler/records/proto",
    protos = [":records_proto"],
    visibility = ["//visibility:public"],
)

go_library(
    name = "records",
    srcs = ["records.go"],
    importpath = "chronicler/records",
    visibility = ["//visibility:public"],
    deps = [
        ":records_go_proto",
        "//web:htmlparser",
        "@commons//collections",
        "@commons//common",
    ],
)

go_test(
    name = "records_test",
    size = "small",
    srcs = [
        "records_test.go",
    ],
    embed = [
        ":records",
    ],
)
