# Security 

## cli

### Server mode

the server is stateless and does not need to save or access by its own  to any information on vault or other of your services

if  you are trying to launch a server and the binary detects a **vault token** within the environment, it will not run.

### port-forward, admin and secrets

the other cli commands need a vault token to work, this vault token could be taken form vault itself or from the menshend web interface

## Web interface

* it stores the vault token in a secure cookie, that can't be read from js
* csrf protection and same origin policy for the login endpoints
* when the api detect that a request come from the browser also apply the csrf and same origin policy to the requests

## Proxy

* the Vault token is deleted from the headers before redirect the request in all the strategies
