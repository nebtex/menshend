# login 

the only credentials needed to operate the menshend binary is a vault token.

you can obtain the vault token from vault, or the menshend web interface

## web interface

it supports  token, github, radius, ldap, and the user/password auth backends.

> if you plan to use github you need to enable it in vault and also pass the github secret and access key to the server using the menshend **config.yml** file

follow this guide for enable or disable any auth backends: https://www.vaultproject.io/docs/auth/index.html
 
