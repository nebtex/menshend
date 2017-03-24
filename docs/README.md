# Kuper

> Warning: if you plan use kuper be sure to understant vault [link] before use it, kuper is just a small piece of sotwatrr
laverage most of build over the vault features. if you knwo vault, will be easy to start to use kuper. 



Kuper is a programmable auth reverse proxy  backed in harshicop vault that allow to create flexible  ccess policy permission  to your services behind a
 firewall/nats/vpn/vpc et. 
is meant to work with http/websocket ans tcp proxy [see how], with as smart impersonification feature, and the future audit logs.

# use case
kuper is mean to be use as internal tool to give access to your infraestrure to the team member and manahe all the control polices in at same plave, 
if you are using a vpn for thi.

eforceless protect your internal applications, like postgres consul , jupyter notebooks, minio instances, gitbook docs, labda functions, etc. the application is endless

impersonitsc

impersonate withint tole


lets breakdown this.

- programmable proxy

> this mean that you should create the function , the only language suported at the moment is lua.


- auth proxy

> the user will need to be loged in order to access to the service behind kuper

  - github
  - user/password
  - token
  - ldap 
 
 the auth feature is backed by the wonderful hardshic app called Vault, so in order to use kuper properly yoy shoul know aboyt vault, 
 you can create thank to vaul versatiles and flexible  access policies to your services, also the lua script give you and additiona flex layer
 to manage who can access or not to your serrver
 
 
 kuper vs other software:
 
 
 ngxix, trefix, pache, caddy
 
 kuper is not mean to compete with those, server 
 
Kuper don't comes and probably will never have tls support so you should use kuper behind one of the proxies mentioned above.
 
 
 jupyterhub
 
 in the long term the idea of kuper is replace jupyterhub
 
 oauth2_proxy
 
 https://github.com/bitly/oauth2_proxy
 
 
 
 https://github.com/movableink/doorman
 
 kuper is  inspired in dorman, but kuper allow to manage 
 * multiples services
 * is prograble 
 * has impersonigfiction
 
 so kuper is the natural evolution of dorman the only thin in iwhinc doreman still has advantage, is tghat has more wide auth provider support 
 
 
 
 
 
