load("@rules_proto//proto:defs.bzl", "proto_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

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
