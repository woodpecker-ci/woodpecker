#!/bin/sh
DOCS_CHANGED=$(echo "$CI_PIPELINE_FILES" | jq -r '.[]' | grep -c '^docs/docs/' || true)
if [ "$DOCS_CHANGED" -gt 0 ]; then
  echo "✅ OK: docs/docs/ has changes"
  exit 0
fi
NON_CLI=$(echo "$CI_PIPELINE_FILES" | jq -r '.[]' | grep -v '^cli/' | grep -v '^cmd/cli/' | grep -v '^docs/' || true)
if [ -z "$NON_CLI" ]; then
  echo "✅ OK: CLI-only feature, docs are auto-generated"
  exit 0
fi
echo "🚨 ERROR: PR has 'feature' label but no changes in docs/docs/"
echo "Please add documentation for the new feature."
exit 1
