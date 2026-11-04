#!/bin/sh
# Perl の E2E テスト（t/feature）を、ビルドした API サーバーに対して実行する。
set -eu
cd "$(dirname "$0")/.."

bin_dir=$(mktemp -d)
go build -o "$bin_dir/server" ./cmd/server
addr=127.0.0.1:18080
"$bin_dir/server" -addr "$addr" &
pid=$!
trap 'kill "$pid"' EXIT

for _ in $(seq 50); do
  curl -fs -o /dev/null -H 'X-User-Id: ping' "http://$addr/api/me" && break
  sleep 0.1
done

TEST_BASE_URL="http://$addr" prove -Ilocal/lib/perl5 -It/lib -r "${@:-t/feature}"
