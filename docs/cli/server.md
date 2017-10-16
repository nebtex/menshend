### <i class="icon-angle-right"/> server, run, start

##### Description
```Run menshend server```

##### Usage
```menshend server --port {port} --config {config} --address {address}```

#### Options
| Name | shorthand | Env var | Default | Description |
| ------ | ------ | ------ | -----| -----|
| --port | -p | |8787 | Bind port|
| --config | -c |[$MENSHEND_CONFIG_FILE] | | config file |
| --address | -a | | "0.0.0.0" | Bind address  |

#### Example
> Run server with the `config.yml` file configuration

```ssh
menshend server -c config.yml
```

```yml
# config.yml
Uris:
  BaseUrl: http://yourdomain.com
hashKey: yourHashKey
blockKey: yourBlockKey
Space:
  Name: Menshend test server
  Logo: http://images.com/logo.png
  Description: My Menshend test server
```
