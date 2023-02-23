#!/bin/bash
source ./github_token.sh

VERSION=$(cat VERSION)

if [[ $1 == '--dry' ]]; then
echo "Release version $VERSION"
goreleaser release --snapshot --clean
else
git tag -a v$VERSION -m "Release version v$VERSION"
git push origin v$VERSION
goreleaser release --clean
fi