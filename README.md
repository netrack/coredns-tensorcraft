# dnstun

_dnstun_ - enable DNS tunneling detection in the service queries.

[![Build Status][BuildStatus]](https://travis-ci.org/netrack/dnstun)

## Description

This is a [CoreDNS](https://coredns.io) plugin that enabled DNS tunneling
detection within submitted queries. It analyzes payload of the DNS query
and either forward the query to the configured resolver (`8.8.8.8` by default),
or returns refuse code.

With `dnstun` enabled, users are able to detect data exfiltration through DNS
tunnels.

## Syntax

```txt
dnstun {
    runtime  HOST:PORT
    detector DETECTOR:VERSION
    [mapping  forward|reverse]
}
```

* `runtime` specifies the endpoint in `HOST:PORT` format to the remote model
runtime. This runtime should comply with e.g. `tensorcraft` HTTP interface.

* `detector` is a directive to configure detector. Option `forward` instructs
the plugin to treat higher probability in the second element of prediction tuple
as DNS tunnel, while `reverse` tells that first element in the prediction tuple
identifies DNS tunnel.

* `mapping` is an optional directive to instructs plugin how interpret the
response from detector: `forward` treats higher probability in the _second_
element of prediction tuple as DNS tunnel, while `reverse` tells that _first_
element in the prediction tuple identifies DNS tunnel. Default is `forward`.

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
        detector dns_cnn:latest
    }
}
```

## Usage

One of the possible ways to run experimental resolver is to use [docker-compose](https://github.com/docker/compose).
In order to run the environment, simply clone this repository and run the following
command:
```sh
% git clone git@github.com:netrack/dnstun.git
% docker-compose up
```

After that, resolver will be accessible at port `53`:
```sh
% dig @localhost google.com
% dig @localhost q+aJ3on2BA.hidemyself.org.
```

[BuildStatus]: https://travis-ci.org/netrack/dnstun.svg?branch=master
