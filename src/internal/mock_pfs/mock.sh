#!/bin/sh
set -e

# Normally, we'd use `mockgen -destination=` directly from
# go:generate, but right now we need a workaround for
# https://github.com/golang/mock/issues/4

dst=apiclient_mock.go
tmp="$dst.tmp"

mockgen github.com/pachyderm/pachyderm/src/pfs APIClient \
	| sed 's,".*/vendor/,",' \
	>"$tmp"
mv -- "$tmp" "$dst"
