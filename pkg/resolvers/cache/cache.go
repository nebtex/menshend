package resolvers

import (
    . "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    . "github.com/nebtex/menshend/pkg/resolvers"
    "github.com/patrickmn/go-cache"
    "time"
    vault "github.com/hashicorp/vault/api"
)

var subdomainCache *cache.Cache

func init() {
    subdomainCache = cache.New(1 * time.Hour, 10 * time.Minute)
}

type CacheResolver struct {
    service *AdminServiceResource
}

func (cr *CacheResolver) SetRequest(method string, body string) {
    cr.service.Resolver.Get().SetRequest(method, body)
    
}
func (cr *CacheResolver) NeedBody() bool {
   return cr.service.Resolver.Get().NeedBody()
}

func NewCacheResolver(service *AdminServiceResource) *CacheResolver {
    return &CacheResolver{service}
}

func (cr *CacheResolver)Resolve(tokenData *vault.Secret) (Backend) {
    if (cr.service.Cache != nil && cr.service.Cache.TTL > 0 &&  tokenData.Data != nil&&
        tokenData.Data["display_name"] != nil && tokenData.Data["display_name"].(string) != "" &&
        tokenData.Data["id"] != nil && tokenData.Data["id"].(string) != "") {
        cacheKey := cr.service.Meta.SubDomain + "_" + tokenData.Data["id"].(string) + "_" + tokenData.Data["display_name"].(string) + "_" + cr.service.Meta.RoleID
        mv, ok := subdomainCache.Get(cacheKey)
        if ok {
            backend := mv.(*SimpleBackend)
            return backend
        }
        bi := cr.service.Resolver.Get().Resolve(tokenData)
        subdomainCache.Set(cacheKey, bi, time.Duration(cr.service.Cache.TTL) * time.Second)
        return bi
        
    } else {
        return cr.service.Resolver.Get().Resolve(tokenData)
    }
}
