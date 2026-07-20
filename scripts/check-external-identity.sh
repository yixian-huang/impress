#!/usr/bin/env bash
# Read-only verification for the external Inkless identity cutover.
set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT" || exit 1

readonly DOMAIN="${INKLESS_DOMAIN:-inkless.run}"
readonly NEW_REPOSITORY="${INKLESS_REPOSITORY:-yixian-huang/inkless}"
readonly OLD_REPOSITORY="${INKLESS_LEGACY_REPOSITORY:-yixian-huang/impress}"
readonly DNS_RESOLVER="${INKLESS_DNS_RESOLVER:-8.8.8.8}"

expect_cutover=false
if [[ "${1:-}" == "--expect-cutover" ]]; then
  expect_cutover=true
elif [[ -n "${1:-}" && "${1:-}" != "--status" ]]; then
  echo "Usage: $0 [--status|--expect-cutover]" >&2
  exit 2
fi

failures=0

record_failure() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

http_status() {
  local status
  status="$(curl --max-time 12 --silent --show-error --output /dev/null \
    --write-out '%{http_code}' "$1" 2>/dev/null)" || status="000"
  printf '%s' "$status"
}

dns_values() {
  dig +time=3 +tries=1 +short "@${DNS_RESOLVER}" "$1" "$2" 2>/dev/null \
    | sed '/^;/d' \
    | sed '/^$/d' \
    | paste -sd, -
}

echo "checked_at_utc=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo "mode=$([[ "$expect_cutover" == true ]] && echo expect-cutover || echo status)"

origin_url="$(git remote get-url origin 2>/dev/null || true)"
echo "git_origin=${origin_url:-missing}"

new_repo_status="$(http_status "https://api.github.com/repos/${NEW_REPOSITORY}")"
old_repo_status="$(http_status "https://api.github.com/repos/${OLD_REPOSITORY}")"
echo "github_new_repository_status=${new_repo_status}"
echo "github_legacy_repository_status=${old_repo_status}"

apex_a="$(dns_values "$DOMAIN" A)"
apex_aaaa="$(dns_values "$DOMAIN" AAAA)"
www_a="$(dns_values "www.${DOMAIN}" A)"
www_aaaa="$(dns_values "www.${DOMAIN}" AAAA)"
www_cname="$(dns_values "www.${DOMAIN}" CNAME)"
nameservers="$(dns_values "$DOMAIN" NS)"
echo "dns_nameservers=${nameservers:-none}"
echo "dns_apex_a=${apex_a:-none}"
echo "dns_apex_aaaa=${apex_aaaa:-none}"
echo "dns_www_a=${www_a:-none}"
echo "dns_www_aaaa=${www_aaaa:-none}"
echo "dns_www_cname=${www_cname:-none}"

https_status="$(http_status "https://${DOMAIN}")"
https_www_status="$(http_status "https://www.${DOMAIN}")"
echo "https_apex_status=${https_status}"
echo "https_www_status=${https_www_status}"

if [[ -n "$apex_a" || -n "$apex_aaaa" ]]; then
  certificate="$({
    openssl s_client -connect "${DOMAIN}:443" -servername "$DOMAIN" </dev/null 2>/dev/null \
      | openssl x509 -noout -subject -issuer -dates -ext subjectAltName 2>/dev/null
  } || true)"
  if [[ -n "$certificate" ]]; then
    echo "tls_certificate=present"
    printf '%s\n' "$certificate" | sed 's/^/tls_detail=/'
  else
    echo "tls_certificate=missing"
  fi
else
  echo "tls_certificate=unavailable-without-address-record"
fi

npm_unscoped_status="$(http_status 'https://registry.npmjs.org/inkless')"
npm_web_status="$(http_status 'https://registry.npmjs.org/%40inkless%2Fweb')"
echo "npm_inkless_status=${npm_unscoped_status}"
echo "npm_inkless_web_status=${npm_web_status}"

module_path="$(awk '$1 == "module" { print $2; exit }' backend/go.mod)"
module_tags="$(git tag --list 'backend/v*' | paste -sd, -)"
go_proxy_status="$(http_status 'https://proxy.golang.org/github.com/yixian-huang/inkless/backend/@v/list')"
echo "go_module=${module_path:-missing}"
echo "go_module_tags=${module_tags:-none}"
echo "go_proxy_status=${go_proxy_status}"

if [[ "$expect_cutover" == true ]]; then
  [[ "$origin_url" == "https://github.com/${NEW_REPOSITORY}.git" ]] \
    || record_failure "origin is not the canonical Inkless repository"
  [[ "$new_repo_status" == "200" ]] \
    || record_failure "canonical GitHub repository is not reachable"
  [[ -n "$apex_a" || -n "$apex_aaaa" ]] \
    || record_failure "apex domain has no A or AAAA record"
  [[ -n "$www_a" || -n "$www_aaaa" || -n "$www_cname" ]] \
    || record_failure "www domain has no address or CNAME record"
  [[ "$https_status" =~ ^(200|204|301|302|307|308)$ ]] \
    || record_failure "apex HTTPS endpoint is not ready"
  [[ "$https_www_status" =~ ^(200|204|301|302|307|308)$ ]] \
    || record_failure "www HTTPS endpoint is not ready"
  [[ "$module_path" == "github.com/yixian-huang/inkless/backend" ]] \
    || record_failure "go.mod does not use the canonical module path"
fi

if (( failures > 0 )); then
  echo "external_identity_ready=false"
  exit 1
fi

if [[ "$expect_cutover" == true ]]; then
  echo "external_identity_ready=true"
else
  echo "external_identity_ready=not-asserted"
fi
