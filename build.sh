#!/bin/bash
set -e

VERSION="1.0.4"
PROTECTED_MODE="no"

export GO15VENDOREXPERIMENT=1

cd $(dirname "${BASH_SOURCE[0]}")
OD="$(pwd)"
WD=$OD

package() {
	echo Packaging $1 Binary
	BUILD_DIR=diagram-${VERSION}-$2-$3
	rm -rf packages/$BUILD_DIR && mkdir -p packages/$BUILD_DIR
	GOOS=$2 GOARCH=$3 ./build.sh
    mv diagram packages/$BUILD_DIR
	cp README.md packages/$BUILD_DIR
	cd packages
	if [ "$2" == "linux" ]; then
		tar -zcf $BUILD_DIR.tar.gz $BUILD_DIR
	else
		zip -r -q $BUILD_DIR.zip $BUILD_DIR
	fi
	rm -rf $BUILD_DIR
	cd ..
}

if [ "$1" == "package" ]; then
	rm -rf packages/
	package "Linux" "linux" "amd64"
	package "Mac" "darwin" "amd64"
	exit
fi

# temp directory for storing isolated environment.
TMP="$(mktemp -d -t sdb.XXXX)"
rmtemp() {
	rm -rf "$TMP"
}
trap rmtemp EXIT

if [ "$NOCOPY" != "1" ]; then
	# copy all files to an isolated directory.
	WD="$TMP/src/github.com/esimov/diagram"
	export GOPATH="$TMP"
	for file in `find . -type f`; do
		if [[ "$file" != "." && "$file" != ./.git* && "$file" != ./diagram ]]; then
			mkdir -p "$WD/$(dirname "${file}")"
			cp -P "$file" "$WD/$(dirname "${file}")"
		fi
	done
	cd $WD
fi

# build and store objects into original directory.
go build -ldflags "-X main.version=$VERSION" -o "$OD/diagram" *.go