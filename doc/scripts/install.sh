#!/bin/bash

PKGDIR=$(dirname "$0")/../../

INSTDIR=/usr/local/bin
DATADIR=/usr/local/share/BookmarkWarrior

mkdir -p "$DATADIR"
cp "$PKGDIR/BookmarkWarrior" "$INSTDIR/"
cp -r "$PKGDIR/tmpl" "$DATADIR/"
cp -r "$PKGDIR/static" "$DATADIR/"
cp "$PKGDIR/Config.toml" "$DATADIR/"
