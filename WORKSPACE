workspace(name = "com_github_lanseg_chronist")

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "com_google_protobuf",
    sha256 = "d0f5f605d0d656007ce6c8b5a82df3037e1d8fe8b121ed42e536f569dec16113",
    strip_prefix = "protobuf-3.14.0",
    urls = [
        "https://mirror.bazel.build/github.com/protocolbuffers/protobuf/archive/v3.14.0.tar.gz",
        "https://github.com/protocolbuffers/protobuf/archive/v3.14.0.tar.gz",
    ],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

git_repository(
    name = "io_bazel_rules_go",
    branch = "master",
    remote = "https://github.com/bazelbuild/rules_go.git",
)

git_repository(
    name = "rules_proto",
    branch = "main",
    remote = "https://github.com/bazelbuild/rules_proto.git",
)

git_repository(
    name = "commons",
    branch = "main",
    remote = "https://github.com/lanseg/golang-commons.git",
)

git_repository(
    name = "tgbot",
    branch = "main",
    remote = "https://github.com/lanseg/tgbot",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

go_rules_dependencies()

go_register_toolchains(version = "1.19.1")
