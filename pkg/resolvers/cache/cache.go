package resolvers

import (
    . "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    . "github.com/nebtex/menshend/pkg/resolvers"
    "github.com/patrickmn/go-cache"
    "time"
    vault "github.com/hashicorp/vault/api"
    "github.com/nebtex/menshend/pkg/config"
)

var subdomainCache *cache.Cache

func init() {
    subdomainCache = cache.New(1 * time.Hour, 10 * time.Minute)
}

type CacheResolver struct {
    service *AdminServiceResource
}

func NewCacheResolver(service *AdminServiceResource) *CacheResolver {
    return &CacheResolver{service}
}

func (cr *CacheResolver)Resolve(tokenData *vault.Secret) (Backend) {
    if (cr.service.Resolver.Cache != nil && cr.service.Resolver.Cache.Active &&   tokenData.Data != nil&&
        tokenData.Data["display_name"] != nil && tokenData.Data["display_name"].(string) != "" &&
        tokenData.Data["id"] != nil && tokenData.Data["id"].(string) != "") {
        cacheKey := cr.service.Meta.SubDomain + "_" + tokenData.Data["id"].(string) + "_" + tokenData.Data["display_name"].(string) + "_" + cr.service.Meta.RoleID
        mv, ok := subdomainCache.Get(cacheKey)
        if ok {
            backend := mv.(*SimpleBackend)
            return backend
        }
        bi := cr.service.Resolver.Get().Resolve(tokenData)
        ttl := cr.service.Resolver.Cache.TTL
        if ttl == 0 {
            ttl = config.Config.DefaultTTL
        }
        subdomainCache.Set(cacheKey, bi, time.Duration(ttl) * time.Second)
        return bi
        
    } else {
        return cr.service.Resolver.Get().Resolve(tokenData)
    }
}
