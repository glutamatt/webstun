# Web Socket Tunnel 

## Expose HTTP Service via a websocke tunnel

Use case: access a webservice behind a private network through a public edge

[ Browser ] --(http request)--> [ Public Edge Server ] --(ws message)--> [ Private reverse proxy ] --(http request)--> [ Private Web Service ]

### TODO

- [ ] Clean Code
- [ ] Factorise Code
- [ ] Add flags
- [ ] Dockerize with docker hub
- [ ] Build Packages with differents architectures (Circle ci ?? Github as artifact repo)
- [ ] Deploy on cloud and run on rasp
