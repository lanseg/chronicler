#!/usr/bin/bash
set -eio pipefail

find -name BUILD -print -exec buildifier {} \;
find -iname '*go' -print \
  -exec gofmt -s -w {} \; \
  -exec goimports -l -w {} \;

bazel test --nocache_test_results --test_output=streamed  //...:all

if [ $? -ne 0 ]; then
 echo “unit tests failed”
 exit 1
fi

