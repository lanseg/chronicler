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
    compilers = ["@rules_go//proto:go_grpc"],
    importpath = "chronicler/storage/endpoint_go_proto",
    protos = [":endpoint_proto"],
    deps = [
        "//records:records_go_proto",
    ],
)

go_binary(
    name = "main",
    srcs = [
        "main.go",
    ],
    deps = [
        ":endpoint",
        ":storage_endpoint_go_proto",
        "//downloader",
        "//records:records_go_proto",
        "//storage",
        "//webdriver",
        "@golang-commons//common",
        "@org_golang_google_grpc//:grpc",
    ],
)

go_library(
    name = "endpoint",
    srcs = [
        "client.go",
        "server.go",
    ],
    importpath = "chronicler/storage/endpoint",
    deps = [
        ":storage_endpoint_go_proto",
        "//records:records_go_proto",
        "//storage",
        "@golang-commons//common",
        "@org_golang_google_grpc//:grpc",
        "@org_golang_google_grpc//credentials/insecure",
    ],
)

go_test(
    name = "endpoint_test",
    size = "small",
    srcs = [
        "endpoint_test.go",
    ],
    embed = [
        ":endpoint",
    ],
    deps = [
        "//downloader",
        "//webdriver",
        "@golang-commons//concurrent",
        "@golang-commons//optional",
    ],
)
