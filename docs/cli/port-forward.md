### <i class="icon-angle-right"/>port-forward</i>

##### Description

```Create a secure tunnel over websockets```

##### Usage

```menshend port-forward --server {server} --role {role} --port {port}```

#### Options

| Flag        | shorthand | Environmental Variable | Description |
| ------      | ------    | --- | -----|
| server    |  s       |  PORT_FORWARD_ENDPOINT   | Full http(s) url of the service under the Menshend space wanted to be tunneled, ip addresses are not allowed |
| role    |  r       |  MD_ROLE  | service role |
| port      |  p       |     | [local-host]:local-port |
| token     |  t       | VAULT_TOKEN   | vault token |
| keepalive |  k       |     | An optional keepalive interval. Since the underlying transport is HTTP, in many instances we'll be traversing through proxies, often these proxies will close idle connections. You must specify a time with a unit, for example '30s' or '2m'. Defaultsto '0s' (disabled) (default: 0s) |
| verbose   |  v       |     | Verbose debug |

##### Examples

> Tunneling a mongo database, locate in the production environment of the example.com organization to the localhost

* make mongo available on `localhost:27017`

```ssh
menshend port-forward --server https://mongo.ml-lab.example.com --role production --port 27017
```
