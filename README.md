Bookmark Warrior
================

[Bookmark Warrior](https://bookmarkwarrior.com/) is a tool to help you collect
and organize the interesting articles and links you find Online (read about it
[on my
website](https://wesleycoakley.com/Projects/bookmark-warrior-organizational-manager.html)!)


Executive Summary
-----------------

With a focus on information density and readability, Bookmark Warrior is an
ad-free and user-supported read-it-later service designed to be available across
all your devices; we help we help you keep track of your favorite online
articles and current news stories so you don't have to!

Account Creation / Activation
-----------------------------

Bookmark Warrior takes advantage of the PayPal API to process account
activations; activation requires a one-time fee which aims to deter bots
from signing up and abusing this service (this is far more effective than
reCaptcha as it turns out).

If you plan to run this service you will need to set up a PayPal account to
process activation transactions and enter your client / secret credentials into
the [PayPal] section of `Config.toml`.

Installation
------------

Create a PayPal to process account activation transactions and ensure you have
Go installed on your machine, then do the following:

```
go get github.com/wesleycoakley/BookmarkWarrior
cd $(go env GOPATH)/src/github.com/wesleycoakley/BookmarkWarrior
cp Config.example.toml Config.toml
# Edit config as necessary
# ...
go build
bash doc/scripts/install.sh
```

... this should install BookmarkWarrior globally to your machine. When you are
ready to run the server, just run the `BookmarkWarrior` binary.

License
-------

Wesley Coakley <w@wesleycoakley.com>

All code is licensed under the MIT License :smile: (see: `/LICENSE`)

