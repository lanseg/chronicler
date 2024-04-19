package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")
load("@rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "status_proto",
    srcs = ["status.proto"],
)

go_proto_library(
    name = "status_go_proto",
    compilers = ["@rules_go//proto:go_grpc"],
    importpath = "chronicler/status/status_go_proto",
    protos = [":status_proto"],
)

go_binary(
    name = "statusserver",
    srcs = [
        "statusserver.go",
    ],
    deps = [
        ":status",
        ":status_go_proto",
        "@golang-commons//common",
        "@org_golang_google_grpc//:grpc",
    ],
)

go_library(
    name = "status",
    srcs = [
        "client.go",
        "server.go",
    ],
    importpath = "chronicler/status",
    deps = [
        ":status_go_proto",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
        "@org_golang_google_grpc//:grpc",
        "@org_golang_google_grpc//credentials/insecure",
        "@org_golang_google_grpc//encoding/gzip",
        "@org_golang_google_grpc//keepalive",
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
    deps = [
        "@golang-commons//concurrent",
        "@golang-commons//optional",
    ],
)
