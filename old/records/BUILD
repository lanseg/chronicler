package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_library", "go_test")
load("@rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "records_proto",
    srcs = ["records.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "records_go_proto",
    importpath = "chronicler/records/proto",
    protos = [":records_proto"],
    visibility = ["//visibility:public"],
)

go_library(
    name = "records",
    srcs = [
        "records.go",
        "sort.go",
    ],
    importpath = "chronicler/records",
    visibility = ["//visibility:public"],
    deps = [
        ":records_go_proto",
        "@golang-commons//almosthtml",
        "@golang-commons//collections",
        "@golang-commons//common",
    ],
)

go_test(
    name = "records_test",
    size = "small",
    srcs = [
        "records_test.go",
        "sort_test.go",
    ],
    embed = [
        ":records",
    ],
)
