#!/usr/bin/env sh
set -eu
go test ./... -count=1 >/tmp/ao-mission-go-test.log
go vet ./... >/tmp/ao-mission-go-vet.log
go build ./cmd/ao-mission
rm -f ao-mission
gofmt -w cmd internal
git diff --check
local_path="/""Users/"
private_key="BEGIN (RSA |OPENSSH |PRIVATE )?PRIVATE ""KEY"
openai_key="sk-""[A-Za-z0-9]{20,}"
github_key="gh[pousr]_""[A-Za-z0-9]{20,}"
secret_assign="tok""en[[:space:]]*[:=][[:space:]]*[^[:space:]]+"
public_safety_pattern="(${local_path}|${private_key}|${openai_key}|${github_key}|${secret_assign})"
for f in README.md docs examples internal cmd scripts; do
  if grep -R -nE "$public_safety_pattern" "$f" >/tmp/ao-mission-public-safety.log 2>/dev/null; then
    echo "public safety scan failed"
    cat /tmp/ao-mission-public-safety.log
    exit 1
  fi
done
echo "AO Mission production readiness: 100/100 status=ready"
