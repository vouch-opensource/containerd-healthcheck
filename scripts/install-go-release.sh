#!/bin/sh
set -e

# fetch and install the latest version from a generic github go release

GH_REPO=$1
NAME=$(echo $GH_REPO | cut -d "/" -f 2)
TAR_FILE="/tmp/$NAME.tar.gz"
RELEASES_URL="https://github.com/$GH_REPO/releases"
CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
DEST="$(dirname $CWD)/.bin"

test -f "$DEST/$NAME" && echo "$NAME already installed!" && exit 0
test -d "$DEST" || mkdir "$DEST"
test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"

last_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" |
    rev |
    cut -f1 -d'/'|
    rev
}

download() {
  test -z "$VERSION" && VERSION="$(last_version)"
  test -z "$VERSION" && {
    echo "Unable to get $NAME version." >&2
    exit 1
  }
  rm -f "$TAR_FILE"
  curl -s -L -o "$TAR_FILE" \
    "${RELEASES_URL}/download/${VERSION}/${NAME}_$(uname -s)_$(uname -m).tar.gz"
}

echo "Fetching $NAME latest version..."
download

echo "Installing $NAME latest version..."
tar -xf "$TAR_FILE" -C "$TMPDIR"
mv "${TMPDIR}/$NAME" .bin/$NAME
chmod +x .bin/$NAME

trap "{ rm -rf $TMPDIR; }" EXIT
