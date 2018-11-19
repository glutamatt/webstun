# WEB Socket TUNnel

[![CircleCI](https://circleci.com/gh/glutamatt/webstun.svg?style=svg)](https://circleci.com/gh/glutamatt/webstun)
[![](https://img.shields.io/badge/docker-glutamatt/webstun-green.svg?logo=docker&longCache=true&style=flat-square)](https://hub.docker.com/r/glutamatt/webstun/)

## Expose HTTP Service via a websocket tunnel

Use case: access a webservice behind a private network through a public edge without router port forwaring configuration

```
+-------------------+      2: http req        +--------------------------+
|                   |  +------------------->  |         webstun          |          Internet
|     Browser       |      7: http res        |    Public Edge Server    |
|                   |  <-------------------+  |                          |
+-------------------+                         +-+----^-----------------^-+
                                                |    |                 |
                                                | 6: ws forward http res
               +-------------------------------------+-----------------+--------------------------+
                                                |    |                 |
                                       3: ws forward http req        1:|websocket connection
                                                +    |                 |
+-------------------------+  5: http res     +--v----+-----------------+-+
|                         +------------------>         webstun           |          Private Network
|    Private Web service  |  4: http req     |    Private Reverse Proxy  |
|                         <------------------+                           |
+-------------------------+                  +---------------------------+

```

### TODO

- [x] Clean Code, Logs, ...
- [ ] Configuration from env variables
- [ ] Secure channels io with timeouts
- [ ] Reconnect the client on error
- [ ] Single backend connected by websocket
- [ ] Secure proxy Access with KEYS
- [ ] Factorise Code
- [x] Mono binary
- [x] Add flags
- [x] Dockerize with docker hub
- [ ] Build Packages with differents architectures (Circle ci ?? Github as artifact repo)
- [ ] Deploy on cloud and run on rasp
