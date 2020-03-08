# dnstun

_dnstun_ - enable DNS tunneling detection in the service queries.

[![Build Status][BuildStatus]](https://travis-ci.org/netrack/dnstun)

## Description

With `dnstun` enabled, users are able to detect data exfiltration through DNS
tunnels.

## Syntax

```txt
dnstun {
    runtime  HOST:PORT
    detector forward|reverse DETECTOR:VERSION
}
```

* `runtime` specifies the endpoint in `HOST:PORT` format to the remote model
runtime. This runtime should comply with e.g. `tensorcraft` HTTP interface.

* `detector` is a directive to configure detector. Option `forward` instructs
the plugin to treat higher probability in the second element of prediction tuple
as DNS tunnel, while `reverse` tells that first element in the prediction tuple
identifies DNS tunnel.

## Examples

Here are the few basic examples of how to enable DNS tunnelling detection.
Usually DNS tunneling detection is turned only for all DNS queries.

Analyze all DNS queries through remote resolver listening on TCP socket.
```txt
.  {
    dnstun {
        # Connect to the runtime that stores model and executes it.
        runtime 10.240.0.1:5678

        # Choose detector and it's version.
        detector reverse dns_cnn:latest
    }
}
```

[BuildStatus]: https://travis-ci.org/netrack/dnstun.svg?branch=master
