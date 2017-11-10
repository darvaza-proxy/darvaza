gpp-server
=========

a server application that can act as a http/https server and act as a proxy server at the same time

Usage
=====

use `gpp-server -dumpflags > server.ini` generate a sample configure file and edit

use `gpp-server -config server.ini` to run it

Notes
=====

if you special the SSL/TLS certificate and private key file, the gpp-server support http2,
    and also support http proxy over http2(TLS)

create a file named `proxy.pac`, the content like this

```
function FindProxyForUrl(url, host){
    return "HTTPS server_name:port";
}
```

replace the `server_name` and `port` to the real one.

configure chrome or firefox to use this proxy.pac, and open a web site to test it

only chrome and firefox support connect to proxy server over TLS, you can use gpp-local to convert the http2 proxy to normal http proxy.
