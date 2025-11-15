#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-}"
if [[ -z "${VERSION}" ]]; then
  echo "Usage: $0 vX.Y.Z"
  exit 1
fi

export GITHUB_REF="refs/tags/${VERSION}"
version="${GITHUB_REF#refs/tags/}"
plain="${version#v}"

# Resolve CHANGELOG.md location (repo root > script dir > current dir)
repo_root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -n "$repo_root" && -f "$repo_root/CHANGELOG.md" ]]; then
  CHANGELOG_FILE="$repo_root/CHANGELOG.md"
elif [[ -f "$script_dir/CHANGELOG.md" ]]; then
  CHANGELOG_FILE="$script_dir/CHANGELOG.md"
elif [[ -f "CHANGELOG.md" ]]; then
  CHANGELOG_FILE="CHANGELOG.md"
else
  CHANGELOG_FILE=""
fi

# Decide where to save RELEASE_NOTES.md
if [[ -n "$CHANGELOG_FILE" ]]; then
  NOTES_FILE="$(dirname "$CHANGELOG_FILE")/RELEASE_NOTES.md"
else
  NOTES_FILE="RELEASE_NOTES.md"
fi

awk -v ver="$plain" '
  BEGIN { in_section=0 }
  $0 ~ "^##[[:space:]]*(\\[)?(v?" ver ")(\\])?([[:space:]]|$)" { in_section=1; next }
  in_section && $0 ~ "^##[[:space:]]" { exit }
  in_section { print }
' "${CHANGELOG_FILE:-/dev/null}" > "$NOTES_FILE" || true

if [[ ! -s "$NOTES_FILE" ]]; then
  echo "No section found for ${version}; writing placeholder."
  echo "Changes for ${version}" > "$NOTES_FILE"
fi

echo "Preview of release notes:"
sed -n '1,80p' "$NOTES_FILE"