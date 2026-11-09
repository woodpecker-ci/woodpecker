#!/bin/bash
shopt -s globstar
shopt -s nullglob

go_files=(**/*.go)
go_files_str=$(IFS=" "; echo "${go_files[*]}")

r=0
for f in "${go_files[@]}"; do
  [[ $f == *_test.go && $f != *fuzz_test.go && $f != *_utils_test.go && $f != e2e/* ]] || continue
  base=${f/_bench_test/}
  base=${base/_linux_test/}
  base=${base/_test/}
  echo "$go_files_str" | grep -Fqw "$base" && continue
  echo "$f"; r=1
done
exit $r
