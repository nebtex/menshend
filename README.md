## menshend

[![codecov](https://codecov.io/gh/nebtex/menshend/branch/master/graph/badge.svg)](https://codecov.io/gh/nebtex/menshend)

* [Full feature list](#download)
* [Docs](#download)
* [Download](#download)
* [Docker](#docker) 


menshend is an identity-aware proxy **(tcp/http)** that uses [Vault](https://github.com/hashicorp/vault) as policy manager, you can use it as replacement of vpns, firewall rules and give access  programmatically to organisation's members, scripts, external users or third party applications.

menshend  was build with the objective of make easy the creation of `secure laboratories`, facilitating the life of **devops/cloud admin** people to whom this product is oriented. :warning: also in order to use effectively menshend you need to already know vault.

a brief list of thing that you can protected

 * organisation internal applications (on-house or open-source)   
 * serverless functions 
 * connect to applications (postgres, redshift, etc.) across diferent vpc on aws, without the need of a vpn, vpc peering, etc. 
 * secure external app for small or medium size sites.
 * give secure access to script, other machines, third party application, web-hooks, on-house slack bots, etc.
 * deploy to kubernetes in a controlled and secure way from your ci pipeline
 * and much more ..

[similar software](#sds) 

[some limitations](#wadas)

## Supported os

|  Operating system | Status |
| --- | --- |
| Linux | [![Build Status](https://travis-ci.org/nebtex/menshend.svg?branch=master)](https://travis-ci.org/nebtex/menshend)|
| Windows | [![Build status](https://ci.appveyor.com/api/projects/status/q8fewu4op9cyxgd5/branch/master?svg=true)](https://ci.appveyor.com/project/criloz/menshend/branch/master)|
| OSX | [![Build Status](https://travis-ci.org/nebtex/menshend.svg?branch=master)](https://travis-ci.org/nebtex/menshend)|

## download

## docker


## Thanks 

vault
vulcan
chisel
kubernetes
swagger


## Licensing

*menshend* is licensed under the Apache License, Version 2.0. See [LICENSE]() for the full license text.

