# dnstun

_dnstun_ - enable DNS tunneling detection in the service queries.

[![Build Status][BuildStatus]](https://travis-ci.org/netrack/dnstun)

## Description

With `dnstun` enabled, users are able to detect data exfiltration through DNS
tunnels.

## Syntax

```txt
dnstun {
    graph PATH
}
```

* `graph` is a directive to configure detector. It is a path to the `.pb` file
with constant graph used to classify DNS traffic.

## Examples

Here are the few basic examples of how to enable DNS tunnelling detection.
Usually DNS tunneling detection is turned only for all DNS queries.

```txt
.  {
    dnstun {
        graph /var/dnstun/dnscnn.pb
    }
}
```

[BuildStatus]: https://travis-ci.org/netrack/dnstun.svg?branch=master
