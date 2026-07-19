#!/usr/bin/env bash
# Reject legacy product brands outside documented, tested compatibility surfaces.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

readonly LEGACY_PATTERN='Impress|impress|IMPRESS|Blotting|blotting|印迹'
readonly ALLOWLIST='scripts/brand-residue-allowlist.txt'
readonly SCAN_PATHS=(
  frontend backend docs-site docs ops scripts test .github
  .env.example .env.sqlite.example
  package.json pnpm-lock.yaml Makefile Dockerfile
  docker-compose.yml docker-compose.sqlite.yml nginx.conf
  README.md CONTRIBUTING.md OPS.md CLAUDE.md .gitignore
)

if [[ ! -f "$ALLOWLIST" ]]; then
  echo "ERROR: missing brand residue allowlist: $ALLOWLIST" >&2
  exit 2
fi

is_allowed() {
  local path="$1"
  local text="$2"
  local path_pattern content_pattern reason

  while IFS='|' read -r path_pattern content_pattern reason; do
    [[ -z "$path_pattern" || "$path_pattern" == \#* ]] && continue
    if [[ "$path" =~ $path_pattern && "$text" =~ $content_pattern ]]; then
      printf '%s' "$reason"
      return 0
    fi
  done < "$ALLOWLIST"

  return 1
}

violations=()
while IFS= read -r match; do
  if [[ ! "$match" =~ ^([^:]+):([0-9]+):([0-9]+):(.*)$ ]]; then
    violations+=("$match")
    continue
  fi

  path="${BASH_REMATCH[1]}"
  text="${BASH_REMATCH[4]}"
  if ! reason="$(is_allowed "$path" "$text")"; then
    violations+=("$match")
  else
    printf 'ALLOW %s (%s)\n' "$match" "$reason"
  fi
done < <(
  rg --hidden --no-heading --line-number --column --color never \
    --glob '!frontend/out/**' \
    --glob '!docs-site/.vitepress/dist/**' \
    --glob '!**/node_modules/**' \
    --glob '!**/coverage/**' \
    --glob '!**/*.zip' \
    --glob '!.git/**' \
    "$LEGACY_PATTERN" "${SCAN_PATHS[@]}" 2>/dev/null || true
)

if (( ${#violations[@]} > 0 )); then
  echo "ERROR: legacy brand residue found outside the documented allowlist:" >&2
  printf '  %s\n' "${violations[@]}" >&2
  exit 1
fi

echo "OK: legacy brand residue is limited to documented compatibility and history."
