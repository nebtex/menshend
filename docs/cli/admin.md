## <i class="icon-angle-right"/>admin, adminServices

##### Description
```Add/update/delete services```

##### Usage
```menshend admin subcommand [subcommand options] [arguments...]```

#### Subcommands
| Name |  Description |
| ------ | -----|
| get, read  | Return service definition |
| delete, remove, del, eliminate, erase | Delete a service |
| upsert, save, apply, update, put, write, upload, add, replace, create, post | Create or update a service |

#### Subcommand options
| Name |   shorthand | Env var | Description |
| ------ | ------ | -----| ----- |
| --role | -r | [$MD_ROLE] |role/namespace/group of services |
| --subdomain | -s | | service subdomain |
| --token | -t | [$VAULT_TOKEN] | vault token |
| --filename | -f | | Filename, or URL to files that contains the configuration to apply|
| --output value | -o value| [$MD_OUTPUT] |Output format json or yaml |
|--api | -a | [$MD_ADDRESS] | Baseurl of the menshend api |

#### Example

> Create a service with the specified `yourservice.yml` file

```
menshend admin apply -f yourservice.yml
```

```yml
#yourservice.yml
api: http://yourdomain.com/v1
kind: AdminService
spec:
  meta:
    roleId: admin
    # Your service will be accessible on http://yourservice.yourdomain.com
    subDomain: yourservice.
    name: yourservice
    description: web terminal.
    logo: http://yourservice.com/logo/png
    tags:
    - admin
    - staff
    longDescription:
      remote:
        url: https://yourservice.repo/README.md
  resolver:
    yaml:
      content: |
        baseUrl: http://yourserviceaddress:yourserviceport
  strategy:
    proxy: {}
```