# Description

Generates an addressbook for usage in MUA's like [aerc](aerc-mail.org) or [mutt](http://www.mutt.org/) from a maildir folder.

**WIP for a few more days**

Features:
- scans all your email (270k emails in 10 seconds on my machine)
- ranks based on both recency and frequency of addresses
- collects from from, to, cc, bcc fields
- if provided with your own addresses it ranks addresses you explicitly sent to higher
- configurable output via go templates
- uses the most frequent name found for each email


# Installation

The easiest way to install is running:
```
go install github.com/ferdinandyb/maildir-rank-addr@latest
```

# Usage

At the minimum, you need to specify where your maildir formatted email are:

```
maildir-rank-addr --maildir=~/.mail
```

Supported flags:

```
      --addresses strings   comma separated list of your email addresses
      --maildir string      path to maildir folder
      --outputpath string   path to output file
      --template string     output template
```

By default results are output to `$HOME/.cache/maildir-rank-addr/addressbook.tsv"`.

Besides the flags, toml formatted configuration file is also possible. It's first looked for in `$HOME/.config/maildir-rank-addr` and then the current working directory.

Complete example configuration with the default (aerc compatible) template:

```
maildir = "~/.mail"
addresses = [
    "address1@example.com",
    "address2@otherexample.com"
]
outputpath = "~/.mail/addressbook"
template = "{{.Address}}\t{{.Name}}"
```

## Integration

#### aerc

Put something like this in your aerc config:
```
address-book-cmd="ugrep -i -Z --color=never %s $HOME/.cache/maildir-rank-addr/addressbook.tsv"
```

# Acknowledgments

Some functions for parsing email was taken from [aerc](aerc-mail.org).
