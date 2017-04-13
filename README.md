## menshend	门神 
[![GitHub release](http://img.shields.io/github/release/nebtex/menshend.svg?style=flat-square)][release]
[![codecov](https://codecov.io/gh/nebtex/menshend/branch/master/graph/badge.svg)](https://codecov.io/gh/nebtex/menshend)

|  Operating system | Status |
| --- | --- |
| Linux | [![Build Status](https://travis-ci.org/nebtex/menshend.svg?branch=master)](https://travis-ci.org/nebtex/menshend)|
| Windows | [![Build status](https://ci.appveyor.com/api/projects/status/q8fewu4op9cyxgd5/branch/master?svg=true)](https://ci.appveyor.com/project/criloz/menshend/branch/master)|
| OSX | [![Build Status](https://travis-ci.org/nebtex/menshend.svg?branch=master)](https://travis-ci.org/nebtex/menshend)|

#### Links

* **⌘** [Full feature list](#download)
* **?** [Docs](#download)
* **⇩** [Download](#download)
* **⌧** [Docker](#docker) 

#### Resume

menshend is an identity-aware reverse proxy **(tcp/http)** that uses [Vault](https://github.com/hashicorp/vault) as policy manager, you can use it as replacement of VPNs, firewall rules and give access programmatically to organization's members, scripts, external users or third party applications.

menshend was built with the objective of makes easy the creation of `secure laboratories`, facilitating the life of **DevOps/cloud admin** people to whom this product is oriented. :warning: also in order to use effectively you need to already know how to install and operate [Vault](https://github.com/hashicorp/vault).

it also comes with a beautiful and functional ui, that make easy to manage the services, login to them from the browser, share secrets, etc.

#### Brief list of thing that you can protect or do.

 * organization internal applications (in-house or open-source)   
 * serverless functions 
 * connect your applications (postgres, redshift, etc.) across different vpc on aws, without the need of a vpn, vpc peering, etc. 
 * secure external app for small or medium size sites.
 * give secure access to scripts, other machines, third party applications, web-hooks, in-house slack bots.
 * deploy to kubernetes in a controlled and secure way from your ci pipeline (travis, gitlab, circle, drone, etc.)
 * and [much more](#sdsd)..., the usage is endless because this is a programmable proxy

see [similar software](#sds)  and [limitations](#wadas)

## Download


## Docker


## Thanks 

without these projects, menshend would not exist.

- [Vault](https://github.com/hashicorp/vault), that act as the central policy manager

- [Oxy](https://github.com/vulcand/oxy),  the heart of the proxy strategy

- [Chisel](https://github.com/jpillora/chisel), we use a adapted version of chisel for create the secure tunnels (port forward strategy) 

- Kubernetes and swagger, the api and the cli tool are inspired on kubernetes and we wrote the api with swagger

## Contribution

To contribute to this project, see [CONTRIBUTING](CONTRIBUTING).

## RoadMap

at the moment we will be focused on fix small issues, and make the software more stable, development of major features  are freeze till we can rewrite the codebase with [omniql](https://github.com/omniql/omniql)

some of the planned  futures are:

* natively support tls and acme 
* add javascript resolver
* reduce the hits to vault
* distributed cache for the resolvers
* improve the performance and make it viable for protecting any kind of external or user facing app

## Licensing

*menshend* is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.

