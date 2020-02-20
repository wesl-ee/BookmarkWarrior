#!/bin/bash

PKGDIR=$(dirname "$0")/../../

INITDIR=/etc/init.d
INSTDIR=/usr/local/bin
DATADIR=/usr/local/share/BookmarkWarrior

mkdir -p "$DATADIR"
echo "Copying init script to $INITDIR"
cp "$PKGDIR/doc/scripts/openrc-bookmarkwarrior.sh" "$INITDIR/bookmarkwarrior"

echo "Copying binary to $INSTDIR"
cp "$PKGDIR/BookmarkWarrior" "$INSTDIR/"

echo "Copying data to $DATADIR"
cp -r "$PKGDIR/tmpl" "$DATADIR/"
cp -r "$PKGDIR/static" "$DATADIR/"
cp "$PKGDIR/Config.toml" "$DATADIR/"

echo "Done!"
