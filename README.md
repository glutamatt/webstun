# Web Socket Tunnel 

## Expose HTTP Service via a websocket tunnel

Use case: access a webservice behind a private network through a public edge without router port forwaring configuration

```
+-------------------+                         +--------------------------+
|                   |      2: http req        |         webstun          |          Internet
|     Browser       |  +------------------->  |    Public Edge Server    |
|                   |                         |                          |
+-------------------+                         +-+----------------------^-+
                                                |                      |
                                                |                      |
               ---------------------------------+--------------------------------------------------
                                                                       |
                                       3: ws forward http req        1: websocket connection
                                                |                      |
+-------------------------+                  +--v----------------------+-+
|                         |                  |         webstun           |          Private Network
|    Private Web service  |  4: http req     |    Private Reverse Proxy  |
|                         <------------------+                           |
+-------------------------+                  +---------------------------+

```

### TODO

- [ ] Clean Code, Logs, ...
- [ ] Configuration from env variables
- [ ] Secure channels io with timeouts
- [ ] Reconnect the client on error
- [ ] Single backend connected by websocket
- [ ] Secure proxy Access with KEYS
- [ ] Factorise Code
- [ ] Mono binary
- [ ] Add flags
- [ ] Dockerize with docker hub
- [ ] Build Packages with differents architectures (Circle ci ?? Github as artifact repo)
- [ ] Deploy on cloud and run on rasp
