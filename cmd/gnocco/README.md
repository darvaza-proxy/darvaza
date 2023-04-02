# Gnocco
## a small cache of goodness

[![Go Report Card](https://goreportcard.com/badge/darvaza.org/gnocco)](https://goreportcard.com/report/darvaza.org/gnocco)
[![Unlicensed](https://img.shields.io/badge/license-Unlicense-blue.svg)](https://darvaza.org/gnocco/blob/master/UNLICENSE)

Gnocco is a DNS cache with resolver, it is based on the wonderful DNS library from Miek Gieben
[dns](https://github.com/miekg/dns) and it is heavily inspired from the DNS cache implemented
by Qiang Ke [godns](https://github.com/kenshinx/godns).
The resolving part will be a correct resolver using a resolving algorithm similar with the one implemented
in [dnscache](http://cr.yp.to/djbdns/dnscache.html).

Gnocco is in very early stages of development with most of its features not implemented yet.

## Quick start
Since commit [09908c2](https://darvaza.org/gnocco/commit/09908c25aa2acd05b93b68d041ee4959cccf80a7) Gnocco removed the code for running under a certain user.
For now the propper way of running Gnocco is:

0. Keep in mind that Gnocco is NOT production state
1. Create an user and group (gnocco user and group are advised but not mandatory)
2. Create a configuration space (directory) for gnocco (ie. /etc/gnocco)
3. Move gnocco.conf and roots files in the configuration space
4. Review and modify gnocco.conf to suit your needs
5. Create a directory to hold logs (ie. /var/log/gnocco)
6. Move the gnocco binary to /usr/bin
7. Run `sudo setcap cap_net_bind_service=+ep /usr/bin/gnocco` in order to enable Gnocco to listen on ports &lt; 1024 (ie. 53)
8. Itegrate Gnocco with your init system
