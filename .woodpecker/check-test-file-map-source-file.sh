#!/bin/bash
shopt -s globstar
shopt -s nullglob

gofiles=(**/*.go)
gofiles_str=$(IFS=" "; echo "${gofiles[*]}")

r=0
for f in "${gofiles[@]}"; do
  [[ $f == *_test.go && $f != *fuzz_test.go && $f != e2e/* ]] || continue
  base=${f/_test/}
  base=${base/_bench_test/}
  echo "$gofiles_str" | grep -qw "$base" && continue
  echo "$f"; r=1
done
exit $r
