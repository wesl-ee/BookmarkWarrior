#!/bin/bash

PKGDIR=$(dirname "$0")/../../
NEWUSER=bookmarkwarrior
INSTDIR=/usr/local/bin
DATADIR=/usr/local/share/BookmarkWarrior

echo "Creating service user: $NEWUSER"
useradd -Ur -s /sbin/nologin -d "$DATADIR" "$NEWUSER"

mkdir -p "$DATADIR"
if [ -f "/sbin/openrc-run" ]; then
	INITDIR="/etc/init.d"
	echo "Copying openrc init script to $INITDIR"

	cp "$PKGDIR/doc/scripts/openrc-bookmarkwarrior.sh" \
		"$INITDIR/bookmarkwarrior"
elif pidof -s systemd>/dev/null; then
	INITDIR="/lib/systemd/system"
	echo "Copying systemd service to $INITDIR"

	cp "$PKGDIR/doc/scripts/systemd-bookmarkwarrior.service" \
		"$INITDIR/bookmarkwarrior.service"
else
	echo "Could not determine init system; not installing service file"
fi

echo "Copying binary to $INSTDIR"
cp "$PKGDIR/BookmarkWarrior" "$INSTDIR/"

echo "Copying data to $DATADIR"
cp -r "$PKGDIR/tmpl" "$DATADIR/"
cp -r "$PKGDIR/static" "$DATADIR/"
chown -R "$NEWUSER:$NEWUSER" "$DATADIR"
chmod -R o-rwx "$DATADIR"

if [ -f "$DATADIR/Config.toml" ]; then
	echo "Not clobbering existing $DATADIR/Config.toml"
else
	cp -n "$PKGDIR/Config.toml" "$DATADIR/"
fi

echo "Done!"
