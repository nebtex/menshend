package resolvers

import (
    . "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/users"
    "github.com/patrickmn/go-cache"
    "time"
    "fmt"
    "strings"
)

var subdomainCache *cache.Cache

func init() {
    subdomainCache = cache.New(1 * time.Hour, 10 * time.Minute)
}

type CacheResolver struct {
}

func getBackend(c *AdminServiceResource, u *User) (*backendImplementation) {
    if c.ProxyLanguage == "yaml" {
        lr := &YAMLResolve{}
        return lr.Resolve(c, u)
    }
    
    if c.ProxyLanguage == "lua" {
        lr := &LuaResolver{}
        return lr.Resolve(c, u)
    }
    
    panic(InternalError.Append(fmt.Sprintf("service: %s,  is programmed in a language that I don't understand", c.ID)))
}

func (yr *CacheResolver)Resolve(c *AdminServiceResource, u *User) (*backendImplementation) {
    if (c.Cache.Active) {
        cacheKey := c.SubDomain + "_" + u.Menshend.Username + string(u.Menshend.AuthProvider) + strings.Join(u.Menshend.Groups, "_")
        mv, ok := subdomainCache.Get(cacheKey)
        if ok {
            backend := mv.(*backendImplementation)
            return backend
        }
        bi := getBackend(c, u)
        subdomainCache.Set(cacheKey, bi, time.Duration(c.Cache.TTL) * time.Second)
        return bi
        
    } else {
        return getBackend(c, u)
    }
    
}
