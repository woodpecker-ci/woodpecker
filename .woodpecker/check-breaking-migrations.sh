#!/bin/sh
MIGRATIONS_CHANGED=$(echo "$CI_PIPELINE_FILES" | jq -r '.[]' | grep -cxF 'docs/src/pages/migrations.md' || true)
if [ "$MIGRATIONS_CHANGED" -gt 0 ]; then
  echo "✅ OK: docs/src/pages/migrations.md has changes"
  exit 0
fi
echo "🚨 ERROR: PR has 'breaking' label but no changes in docs/src/pages/migrations.md"
echo "Please add a migration note for the breaking change."
exit 1
