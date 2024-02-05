package(default_visibility = ["//visibility:public"])

load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")
load("@rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "endpoint_proto",
    srcs = ["endpoint.proto"],
    deps = ["//records:records_proto"],
)

go_proto_library(
    name = "storage_endpoint_go_proto",
    importpath = "chronicler/storage/endpoint",
    compilers = ["@rules_go//proto:go_grpc"],
    protos = [":endpoint_proto"],
)

go_binary(
    name = "server",
    srcs = ["main.go"],
    deps = [
        ":storage_endpoint_go_proto",
        "//records:records_go_proto",
        "@golang-commons//common",
        "@org_golang_google_grpc//:grpc",
    ],
)

go_library(
    name = "storage",
    srcs = [
        "overlay.go",
        "storage.go",
    ],
    importpath = "chronicler/storage",
    deps = [
        "//downloader",
        "//records",
        "//records:records_go_proto",
        "//util",
        "//webdriver",
        "@golang-commons//collections",
        "@golang-commons//common",
        "@golang-commons//optional",
    ],
)

go_test(
    name = "storage_test",
    size = "small",
    srcs = [
        "overlay_test.go",
        "storage_test.go",
    ],
    embed = [
        ":storage",
    ],
)
