#!/bin/sh
# Scan the container images of all verified Woodpecker plugins with Trivy.
#
# The list of plugins comes from docs/plugins/woodpecker-plugins/plugins.json.
# The image of a plugin is read from the `containerImage` field in the
# frontmatter of the plugin's docs.md file. Images without a tag are scanned as
# `:latest`; the digest recorded in the report is the digest that was actually
# scanned.
#
# This script never aborts the loop because a single plugin could not be
# fetched or scanned. Failures are collected in the report and only the
# summary step decides whether the workflow should be marked as failed.
set -u

PLUGINS_JSON="${PLUGINS_JSON:-docs/plugins/woodpecker-plugins/plugins.json}"
REPORT_DIR="${REPORT_DIR:-plugin-scan-report}"
TRIVY_IMAGE="${TRIVY_IMAGE:-unknown}"
SCANNERS="${SCANNERS:-vuln,secret}"
DB_REPOSITORY="${DB_REPOSITORY:-mirror.gcr.io/aquasec/trivy-db:2}"
SCAN_TIMEOUT="${SCAN_TIMEOUT:-15m}"
FAIL_ON_SCAN_ERROR="${FAIL_ON_SCAN_ERROR:-false}"
WEBSITE_RAW_BASE="${WEBSITE_RAW_BASE:-https://raw.githubusercontent.com/woodpecker-ci/woodpecker-ci.github.io/plugin-scan-results}"

log() {
  printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"
}

die() {
  log "ERROR: $*"
  exit 1
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    printf 'unknown'
  fi
}

normalize_image() {
  case "$1" in
    *@*)
      printf '%s' "$1"
      ;;
    *)
      last="${1##*/}"
      case "$last" in
        *:*)
          printf '%s' "$1"
          ;;
        *)
          printf '%s:latest' "$1"
          ;;
      esac
      ;;
  esac
}

image_has_tag() {
  case "$1" in
    *@*)
      return 0
      ;;
  esac
  last="${1##*/}"
  case "$last" in
    *:*)
      return 0
      ;;
  esac
  return 1
}

# Return the Docker Hub repository path for an unqualified image reference.
# Images from another registry (ghcr.io, quay.io, ...) are not handled here.
dockerhub_repo() {
  case "$1" in
    docker.io/*)
      printf '%s' "${1#docker.io/}"
      ;;
    */*)
      first="${1%%/*}"
      case "$first" in
        *.* | *:* | localhost)
          return 1
          ;;
        *)
          printf '%s' "$1"
          ;;
      esac
      ;;
    *)
      printf '%s' "$1"
      ;;
  esac
}

# Some plugins only publish a `next` tag (or version tags) and no `latest`.
# If scanning `:latest` failed, pick another published tag from Docker Hub.
resolve_alternative_image() {
  image_ref="$1"
  case "$image_ref" in
    *:latest)
      base="${image_ref%:latest}"
      ;;
    *)
      return 1
      ;;
  esac

  repo="$(dockerhub_repo "$base")" || return 1
  tags="$(curl -fsSL --connect-timeout 10 --max-time 30 \
    "https://hub.docker.com/v2/repositories/${repo}/tags?page_size=100&ordering=last_updated" 2>/dev/null \
    | jq -r '.results[]?.name' 2>/dev/null)" || return 1
  [ -n "$tags" ] || return 1

  for candidate in next edge main master; do
    if printf '%s\n' "$tags" | grep -qx "$candidate"; then
      printf '%s:%s' "$base" "$candidate"
      return 0
    fi
  done

  tag="$(printf '%s\n' "$tags" | grep -Ev '^(latest|next|edge|main|master)$' | grep -Ev -- '-[^-]*$' | head -n 1)"
  [ -n "$tag" ] || tag="$(printf '%s\n' "$tags" | grep -vx 'latest' | head -n 1)"
  [ -n "$tag" ] || return 1
  printf '%s:%s' "$base" "$tag"
}

run_trivy_scan() {
  scan_image="$1"
  scan_output="$2"
  scan_log="$3"
  trivy image \
    --image-src remote \
    --scanners "$SCANNERS" \
    --db-repository "$DB_REPOSITORY" \
    --format json \
    --output "$scan_output" \
    --no-progress \
    --timeout "$SCAN_TIMEOUT" \
    --exit-code 0 \
    "$scan_image" >"$scan_log" 2>&1
}

command -v jq >/dev/null 2>&1 || die "jq is required"
command -v trivy >/dev/null 2>&1 || die "trivy is required"
[ -f "$PLUGINS_JSON" ] || die "plugin index not found: $PLUGINS_JSON"
case "$REPORT_DIR" in
  "" | "/" | ".")
    die "refusing to use unsafe report directory: $REPORT_DIR"
    ;;
esac

rm -rf "$REPORT_DIR"
mkdir -p "$REPORT_DIR/reports"

plugin_list="$REPORT_DIR/plugins.tsv"
if ! jq -r '
  .plugins
  | to_entries[]
  | select(.value.verified == true)
  | [(.key | tostring), .value.name, .value.docs]
  | @tsv
' "$PLUGINS_JSON" > "$plugin_list"; then
  die "could not parse $PLUGINS_JSON"
fi

total="$(wc -l < "$plugin_list" | tr -d ' ')"
log "Scanning $total verified plugin images"
log "Trivy image: $TRIVY_IMAGE"
log "Scanners: $SCANNERS"
log "Vulnerability DB repository: $DB_REPOSITORY"

index=0
while IFS="$(printf '\t')" read -r plugin_key name docs_url; do
  [ -n "$name" ] || continue
  index=$((index + 1))

  slug="$(printf '%s' "$name" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9][^a-z0-9]*/-/g; s/^-//; s/-$//')"
  [ -n "$slug" ] || slug="plugin"
  slug="$(printf '%03d-%s' "$index" "$slug")"
  plugin_dir="$REPORT_DIR/reports/$slug"
  details_url="$WEBSITE_RAW_BASE/$slug.json"
  details_file="$REPORT_DIR/website/details/$slug.json"
  mkdir -p "$plugin_dir" "$REPORT_DIR/website/details"

  started_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  start_epoch="$(date +%s)"
  status="ok"
  error=""
  image=""
  image_requested=""
  image_selection="docs"
  image_had_tag="true"
  image_digest=""
  image_digest_ref=""
  image_id=""
  exit_code=0
  findings='{"critical":0,"high":0,"medium":0,"low":0,"unknown":0,"secrets":0,"misconfigurations":0,"total":0}'
  trivy_report="$plugin_dir/trivy.json"

  log "[$index/$total] $name"

  if ! curl -fsSL --retry 3 --retry-delay 2 --connect-timeout 20 --max-time 60 \
      --output "$plugin_dir/docs.md" "$docs_url" 2>"$plugin_dir/docs.err"; then
    status="error"
    error="could not fetch plugin docs: $docs_url"
  fi

  if [ "$status" = "ok" ]; then
    image="$(awk '
      BEGIN { in_frontmatter = 0; done = 0 }
      /^---[[:space:]]*$/ {
        if (in_frontmatter) { done = 1; exit }
        in_frontmatter = 1
        next
      }
      in_frontmatter && !done { print }
    ' "$plugin_dir/docs.md" \
      | sed -n 's/^containerImage:[[:space:]]*//p' \
      | head -n 1 \
      | sed 's/[[:space:]]*$//; s/^"//; s/"$//; s/^'\''//; s/'\''$//')"

    if [ -z "$image" ]; then
      status="error"
      error="no containerImage found in plugin docs frontmatter"
    else
      if image_has_tag "$image"; then
        image_had_tag="true"
      else
        image_had_tag="false"
      fi
      image="$(normalize_image "$image")"
      image_requested="$image"
    fi
  fi

  if [ "$status" = "ok" ]; then
    log "  scanning $image"
    run_trivy_scan "$image" "$trivy_report" "$plugin_dir/trivy.log"
    exit_code=$?

    # Several plugins do not publish a `latest` tag. If the docs image was
    # untagged and `:latest` does not exist, pick a published tag instead.
    if [ "$exit_code" -ne 0 ] && [ "$image_had_tag" = "false" ] \
        && grep -qiE 'MANIFEST_UNKNOWN|manifest unknown|not found|no such manifest' "$plugin_dir/trivy.log"; then
      alternative="$(resolve_alternative_image "$image_requested" 2>/dev/null || true)"
      if [ -n "$alternative" ] && [ "$alternative" != "$image" ]; then
        log "  $image is unavailable, retrying with $alternative"
        mv "$plugin_dir/trivy.log" "$plugin_dir/trivy-latest.log"
        image="$alternative"
        image_selection="dockerhub-fallback"
        run_trivy_scan "$image" "$trivy_report" "$plugin_dir/trivy.log"
        exit_code=$?
      fi
    fi

    if [ "$exit_code" -ne 0 ]; then
      status="error"
      error="trivy exited with code $exit_code"
    fi
  fi

  if [ -s "$trivy_report" ]; then
    image_digest_ref="$(jq -r '.Metadata.RepoDigests[0] // empty' "$trivy_report" 2>/dev/null)"
    image_digest="${image_digest_ref#*@}"
    image_id="$(jq -r '.Metadata.ImageID // empty' "$trivy_report" 2>/dev/null)"
    findings="$(jq -c '
      [.Results[]? | .Vulnerabilities[]?] as $vulns
      | {
          critical: ([$vulns[] | select(.Severity == "CRITICAL")] | length),
          high: ([$vulns[] | select(.Severity == "HIGH")] | length),
          medium: ([$vulns[] | select(.Severity == "MEDIUM")] | length),
          low: ([$vulns[] | select(.Severity == "LOW")] | length),
          unknown: ([$vulns[] | select(.Severity == "UNKNOWN")] | length),
          secrets: ([.Results[]? | .Secrets[]?] | length),
          misconfigurations: ([.Results[]? | .Misconfigurations[]?] | length),
          total: ($vulns | length)
        }
    ' "$trivy_report" 2>/dev/null)" || findings='{"critical":0,"high":0,"medium":0,"low":0,"unknown":0,"secrets":0,"misconfigurations":0,"total":0}'
  fi

  # Detailed, trimmed report used by the plugin detail page. Keep only the
  # fields needed to render the findings table; the raw Trivy report contains
  # megabytes of package metadata that is not useful on the website.
  if [ -s "$trivy_report" ]; then
    details_source="$trivy_report"
  else
    details_source="$plugin_dir/trivy-empty.json"
    printf '{"Results":[]}' > "$details_source"
  fi

  jq \
    --arg name "$name" \
    --arg docs "$docs_url" \
    --arg image "$image" \
    --arg image_digest "$image_digest" \
    --arg image_digest_ref "$image_digest_ref" \
    --arg scanned_at "$started_at" \
    --arg status "$status" \
    --arg error "$error" \
    --argjson findings "$findings" \
    '{
      schema_version: 1,
      name: $name,
      docs: $docs,
      image: (if $image == "" then null else $image end),
      image_digest: (if $image_digest == "" then null else $image_digest end),
      image_digest_ref: (if $image_digest_ref == "" then null else $image_digest_ref end),
      scanned_at: $scanned_at,
      status: $status,
      error: (if $error == "" then null else $error end),
      findings: $findings,
      vulnerabilities: [
        .Results[]?
        | .Target as $target
        | .Vulnerabilities[]?
        | {
            target: $target,
            id: .VulnerabilityID,
            package: .PkgName,
            installed_version: .InstalledVersion,
            fixed_version: (if .FixedVersion == "" then null else .FixedVersion end),
            severity: .Severity,
            title: .Title,
            url: .PrimaryURL
          }
      ],
      secrets: [
        .Results[]?
        | .Target as $target
        | .Secrets[]?
        | {
            target: $target,
            rule_id: .RuleID,
            title: .Title,
            severity: .Severity,
            match: .Match
          }
      ],
      misconfigurations: [
        .Results[]?
        | .Target as $target
        | .Misconfigurations[]?
        | {
            target: $target,
            id: .ID,
            title: .Title,
            severity: .Severity,
            message: .Message,
            resolution: .Resolution
          }
      ]
    }' "$details_source" > "$details_file"

  finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  duration_seconds=$(( $(date +%s) - start_epoch ))

  jq -n \
    --arg name "$name" \
    --arg docs "$docs_url" \
    --arg image "$image" \
    --arg image_requested "$image_requested" \
    --arg image_selection "$image_selection" \
    --arg image_digest "$image_digest" \
    --arg image_digest_ref "$image_digest_ref" \
    --arg image_id "$image_id" \
    --arg started_at "$started_at" \
    --arg finished_at "$finished_at" \
    --argjson duration_seconds "$duration_seconds" \
    --argjson exit_code "$exit_code" \
    --arg status "$status" \
    --arg error "$error" \
    --arg report "$trivy_report" \
    --arg details_url "$details_url" \
    --argjson findings "$findings" \
    '{
      name: $name,
      docs: $docs,
      image: (if $image == "" then null else $image end),
      image_requested: (if $image_requested == "" then null else $image_requested end),
      image_selection: $image_selection,
      image_digest: (if $image_digest == "" then null else $image_digest end),
      image_digest_ref: (if $image_digest_ref == "" then null else $image_digest_ref end),
      image_id: (if $image_id == "" then null else $image_id end),
      started_at: $started_at,
      finished_at: $finished_at,
      duration_seconds: $duration_seconds,
      exit_code: $exit_code,
      status: $status,
      error: (if $error == "" then null else $error end),
      findings: $findings,
      report: $report,
      details_url: $details_url
    }' > "$plugin_dir/meta.json"

  if [ "$status" = "ok" ]; then
    log "  ok: $(printf '%s' "$findings" | jq -r '"critical=\(.critical) high=\(.high) medium=\(.medium) low=\(.low) secrets=\(.secrets)"')"
  else
    log "  FAILED: $error"
  fi
done < "$plugin_list"

# Record exactly which Trivy was used for the scans. The binary checksum makes
# the report verifiable even if the image tag is moved later.
scanner_raw="$REPORT_DIR/trivy-version.json"
if ! trivy version --format json > "$scanner_raw" 2>"$REPORT_DIR/trivy-version.err"; then
  trivy --version > "$scanner_raw" 2>/dev/null || true
fi

trivy_binary="$(command -v trivy)"
trivy_binary_sha="sha256:$(sha256_file "$trivy_binary")"

jq -n \
  --arg image "$TRIVY_IMAGE" \
  --arg binary "$trivy_binary" \
  --arg binary_sha256 "$trivy_binary_sha" \
  --arg scanners "$SCANNERS" \
  --arg db_repository "$DB_REPOSITORY" \
  --slurpfile version "$scanner_raw" \
  '{
    image: $image,
    binary: $binary,
    binary_sha256: $binary_sha256,
    scanners: ($scanners | split(",")),
    db_repository: $db_repository,
    version: ($version[0] // {})
  }' > "$REPORT_DIR/scanner.json"

generated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

jq -s \
  --arg generated_at "$generated_at" \
  --slurpfile scanner "$REPORT_DIR/scanner.json" \
  '
  . as $plugins
  | {
      schema_version: 1,
      generated_at: $generated_at,
      scanner: $scanner[0],
      totals: {
        plugins: ($plugins | length),
        scanned: ([$plugins[] | select(.status == "ok")] | length),
        failed: ([$plugins[] | select(.status != "ok")] | length),
        findings: {
          critical: ([$plugins[].findings.critical // 0] | add // 0),
          high: ([$plugins[].findings.high // 0] | add // 0),
          medium: ([$plugins[].findings.medium // 0] | add // 0),
          low: ([$plugins[].findings.low // 0] | add // 0),
          unknown: ([$plugins[].findings.unknown // 0] | add // 0),
          secrets: ([$plugins[].findings.secrets // 0] | add // 0),
          misconfigurations: ([$plugins[].findings.misconfigurations // 0] | add // 0),
          total: ([$plugins[].findings.total // 0] | add // 0)
        }
      },
      plugins: $plugins
    }
  ' "$REPORT_DIR"/reports/*/meta.json > "$REPORT_DIR/summary.json"

# Compact result file consumed by the plugins website. Keep this small: it is
# published to the website repository on every scan, not committed to this repo.
mkdir -p "$REPORT_DIR/website"
jq \
  --arg pipeline_url "${CI_PIPELINE_URL:-}" \
  '{
    schema_version: 1,
    generated_at: .generated_at,
    pipeline_url: (if $pipeline_url == "" then null else $pipeline_url end),
    scanner: {
      version: (.scanner.version.Version // null),
      binary_sha256: (.scanner.binary_sha256 // null),
      db_version: (.scanner.version.VulnerabilityDB.Version // null),
      db_updated_at: (.scanner.version.VulnerabilityDB.UpdatedAt // null)
    },
    plugins: [
      .plugins[] | {
        docs: .docs,
        name: .name,
        image: .image,
        image_digest: .image_digest,
        image_digest_ref: .image_digest_ref,
        status: .status,
        scanned_at: .started_at,
        duration_seconds: .duration_seconds,
        error: .error,
        details_url: .details_url,
        findings: .findings
      }
    ]
  }' "$REPORT_DIR/summary.json" > "$REPORT_DIR/website/plugin-scan.json"

jq -r '
  "# Plugin image scan",
  "",
  "- Generated: \(.generated_at)",
  "- Scanner image: `\(.scanner.image)`",
  "- Scanner version: \(.scanner.version.Version // "unknown")",
  "- Scanner binary: `\(.scanner.binary)` (\(.scanner.binary_sha256))",
  "- Vulnerability DB: version \(.scanner.version.VulnerabilityDB.Version // "unknown"), updated \(.scanner.version.VulnerabilityDB.UpdatedAt // "unknown")",
  "- Scanners: \(.scanner.scanners | join(", "))",
  "- Plugins: \(.totals.plugins) total, \(.totals.scanned) scanned, \(.totals.failed) failed",
  "- Findings: \(.totals.findings.critical) critical, \(.totals.findings.high) high, \(.totals.findings.medium) medium, \(.totals.findings.low) low, \(.totals.findings.unknown) unknown, \(.totals.findings.secrets) secrets, \(.totals.findings.misconfigurations) misconfigurations",
  "",
  "| Plugin | Image | Image digest | Status | Started (UTC) | Duration | Crit | High | Med | Low | Unk | Secrets |",
  "| --- | --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |",
  (.plugins[] | "| \(.name) | `\(.image // "-")` | `\(.image_digest // "-")` | \(.status)\(if .error then ": " + .error else "" end) | \(.started_at) | \(.duration_seconds)s | \(.findings.critical) | \(.findings.high) | \(.findings.medium) | \(.findings.low) | \(.findings.unknown) | \(.findings.secrets) |"),
  "",
  "Detailed Trivy JSON reports:",
  "",
  (.plugins[] | "- \(.name): `\(.report)`")
' "$REPORT_DIR/summary.json" > "$REPORT_DIR/summary.md"

cat "$REPORT_DIR/summary.md"

failed="$(jq -r '.totals.failed' "$REPORT_DIR/summary.json")"
if [ "$failed" -gt 0 ] && [ "$FAIL_ON_SCAN_ERROR" = "true" ]; then
  log "$failed plugin scan(s) failed"
  exit 1
fi

log "Report written to $REPORT_DIR/summary.json (detailed: $REPORT_DIR/reports/)"
log "Website status written to $REPORT_DIR/website/plugin-scan.json"
