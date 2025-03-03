#!/usr/bin/bash

find -name BUILD -print -exec buildifier {} \;
find -iname '*go' -not -path './chronicler_storage/*' \
  -print \
  -exec gofmt -s -w {} \; \
  -exec goimports -l -w {} \;

find -iname '*go' | xargs grep '"HERE'
if [ $? -eq 0 ]; then
  echo "Debug output found"
  exit 1
fi

find -iname '*css' -o -iname '*js' -o -iname '*html' -not -path '*/testdata/*' | xargs prettier --print-width 100 --tab-width 4 --write
bazel test --enable_bzlmod --nocache_test_results --test_output=streamed  //...:all

if [ $? -ne 0 ]; then
 echo “unit tests failed”
 exit 1
fi


