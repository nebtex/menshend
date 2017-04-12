## menshend

[![codecov](https://codecov.io/gh/nebtex/menshend/branch/master/graph/badge.svg)](https://codecov.io/gh/nebtex/menshend)

* :1234: [Full feature list](#download)
* :page_facing_up: [Docs](#download)
* :arrow_down: [Download](#download)
* :package: [Docker](#docker) 


menshend is an identity-aware proxy **(tcp/http)** that uses [Vault](https://github.com/hashicorp/vault) as policy manager, you can use it as replacement of vpns, firewall rules and give access  programmatically to organisation's members, scripts, external users or third party applications.

menshend  was build with the objective of make easy the creation of `secure laboratories`, facilitating the life of **devops/cloud admin** people to whom this product is oriented. :warning: also in order to use effectively menshend you need to already know how install and operate [Vault](https://github.com/hashicorp/vault).

#### a brief list of thing that you can protected

 * organisation internal applications (in-house or open-source)   
 * serverless functions 
 * connect to applications (postgres, redshift, etc.) across diferent vpc on aws, without the need of a vpn, vpc peering, etc. 
 * secure external app for small or medium size sites.
 * give secure access to script, other machines, third party application, web-hooks, in-house slack bots, etc.
 * deploy to kubernetes in a controlled and secure way from your ci pipeline (travis, gitlab, circle, drone, etc.)
 * and [much more ..](#sdsd)

see [similar software](#sds)  and [some limitations](#wadas)

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

