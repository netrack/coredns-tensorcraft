# dnstun

_dnstun_ - enable DNS tunneling detection in the service queries.

## Description

With `dnstun` enabled, users are able to detect data exfiltration through DNS
tunnels.

## Syntax

```txt
dnstun SOCKET
```

* SOCKET required endpoint to the remote detector.

## Examples

Here are the few basic examples of how to enable DNS tunnelling detection.
Usually DNS tunneling detection is turned only for all DNS queries.

Analyze all DNS queries through remote resolver listening on UNIX socket.
```txt
. {
    dnstun unix:///var/run/dnstun.sock
}
```

Analyze all DNS queries through remote resolver listening on TCP socket.
```txt
.  {
    dnstun tcp://10.240.0.1:5678
}
```
