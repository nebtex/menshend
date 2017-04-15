## menshend	门神 
[![GitHub release](http://img.shields.io/github/release/nebtex/menshend.svg?style=flat-square)][release]
[![codecov](https://codecov.io/gh/nebtex/menshend/branch/master/graph/badge.svg)](https://codecov.io/gh/nebtex/menshend)

[release]: https://github.com/nebtex/menshend/releases


|  Operating system | Status |
| --- | --- |
| Linux | [![Build Status](https://travis-ci.org/nebtex/menshend.svg?branch=master)](https://travis-ci.org/nebtex/menshend)|
| Windows | [![Build status](https://ci.appveyor.com/api/projects/status/q8fewu4op9cyxgd5/branch/master?svg=true)](https://ci.appveyor.com/project/criloz/menshend/branch/master)|
| OSX | [![Build Status](https://travis-ci.org/nebtex/menshend.svg?branch=master)](https://travis-ci.org/nebtex/menshend)|

#### Links

* **⌘** [Full feature list](#download)
* **?** [Docs](#download)
* **⇩** [Download](#binaries)
* **⌧** [Docker](#docker) 

#### Resume

Menshend is an identity-aware reverse proxy **(TCP/HTTP)** that uses [Vault](https://github.com/hashicorp/vault) as policy manager. You can use it as replacement for VPNs, firewall rules and to give access programmatically to organization's members, scripts, external users or third party applications.

Menshend was built with the objective of making the `secure laboratories` creation easy, facilitating the life of **DevOps/cloud admin** engineers, whom this product is oriented to. 

:warning: In order to use it effectively, you already need to know how to install and operate [Vault](https://github.com/hashicorp/vault).

It does also come with a beautiful and functional UI which makes it simple to manage the services, login to them from the browser, share secrets, etc.

#### Brief list of things you can protect or do:

 * Organization internal applications (in-house or open-source).
 * Serverless functions.
 * Connect your applications (PostgreSQL, Redshift, etc.) across different VPCs on AWS, without the need of a VPN, VPC peering, etc. 
 * Secure external APPs for small or medium size sites.
 * Give secure access to scripts, other machines, third party applications, web-hooks, in-house slack bots.
 * Deploy to Kubernetes in a controlled and secure way from your CI pipelines (Travis CI, Gitlab, CircleCI, Drone, etc.).
 * and [much more](#sdsd)..., its usages are endless because of being a programmable proxy.

See [similar software](#sds)  and [limitations](#wadas)


## Binaries

[![Releases](https://img.shields.io/github/downloads/nebtex/menshend/total.svg)][release]

#### OS X 
```shell
curl -LO https://github.com/nebtex/menshend/releases/download/$(curl -s https://raw.githubusercontent.com/nebtex/menshend/master/stable.txt)/menshend_darwin_amd64.zip
```

#### Linux
```shell
curl -LO https://github.com/nebtex/menshend/releases/download/$(curl -s https://raw.githubusercontent.com/nebtex/menshend/master/stable.txt)/menshend_linux_amd64.zip
```

#### Windows

```shell 
curl -LO https://github.com/nebtex/menshend/releases/download/$(curl -s https://raw.githubusercontent.com/nebtex/menshend/master/stable.txt)/menshend_windows_amd64.zip
```

unzip and make the menshend binary executable and move it to your PATH 

full list of downloads for other platforms [here][release]

## Docker

[![](https://images.microbadger.com/badges/image/nebtex/menshend.svg)](https://microbadger.com/images/nebtex/menshend "Get your own image badge on microbadger.com")
[![Docker Pulls](https://img.shields.io/docker/pulls/nebtex/menshend.svg)](https://hub.docker.com/r/nebtex/menshend/)

full list of tags, configurations and options can be found [here](https://hub.docker.com/r/nebtex/menshend/)  

### linux amd64

```shell 
docker pull nebtex/menshend:$(curl -s https://raw.githubusercontent.com/nebtex/menshend/master/stable.txt)
``` 

## Thanks 

Without these projects, menshend would not exist.

- [Vault](https://github.com/hashicorp/vault), as the central policy manager.

- [Oxy](https://github.com/vulcand/oxy), the heart of the proxying strategy.

- [Chisel](https://github.com/jpillora/chisel), we use an adapted version of Chisel to create secured tunnels (port forwarding strategy).

- Kubernetes and Swagger, the API and CLI tools are inspired on Kubernetes, and we implemented the API with Swagger.


## Contribution

To contribute to this project, see [CONTRIBUTING](CONTRIBUTING).

## RoadMap

At the moment we will be focused on fixing small issues and making the software more stable. Development of major features is froze till we can rewrite the codebase with [omniql](https://github.com/nebtex/omniql).

Some of the planned features are:

* Natively support TLS and ACME.
* Add Javascript resolver.
* Reduce the hits to Vault.
* Distributed cache for the resolvers.
* Improve the performance and make it viable for protecting any kind of external or user facing APP.


## Licensing

*menshend* is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.

