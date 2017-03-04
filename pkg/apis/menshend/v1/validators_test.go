package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils/test"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    "github.com/ansel1/merry"
)

func Test_ValidateService(t *testing.T) {
    services := []string{
        "roles/admin/consul.",
        "roles/admin-core/consul.production.",
        "roles/role-co0e/con-sul.prod-uction.",
    }
    
    badServices := []string{
        "roles/admin/consul",
        "roles/admin-core/$consul.production.",
        "roles/role-co0e/service/con-sul.prod-uction.",
    }
    
    Convey("Test ValidateService", t, func() {
        for _, service := range services {
            ValidateService(service)
        }
    })
    
    Convey("Test ValidateService (bad services)", t, func(c C) {
        wraper := func(service string) {
            defer func() {
                r := recover()
                c.So(r, ShouldNotBeNil)
            }()
            ValidateService(service)
        }
        for _, service := range badServices {
            wraper(service)
        }
    })
}

func Test_ValidateRole(t *testing.T) {
    roles := []string{
        "admin",
        "admin-core",
        "role-co0e",
    }
    badRoles := []string{
        "admin.cc",
        "a%dmin-core",
        "r:ole-co0e",
    }
    Convey("Test ValidateRole", t, func() {
        for _, role := range roles {
            ValidateRole(role)
        }
    })
    
    Convey("Test ValidateRole (bad roles)", t, func(c C) {
        wraper := func(role string) {
            defer func() {
                r := recover()
                c.So(r, ShouldNotBeNil)
            }()
            ValidateRole(role)
        }
        for _, role := range badRoles {
            wraper(role)
        }
    })
}

func Test_ValidateSubdomain(t *testing.T) {
    subdomains := []string{
        "consul.",
        "consul.production.",
        "con-sul.prod-uction.",
    }
    
    badSubdomains := []string{
        "con..sul.",
        "consul.production",
        "con-sul.prod-ucti$$on.",
    }
    Convey("Test ValidateSubdomain", t, func() {
        for _, subdomain := range subdomains {
            ValidateSubdomain(subdomain)
        }
    })
    
    Convey("Test ValidateSubdomain (bad subdomains)", t, func(c C) {
        wraper := func(sd string) {
            defer func() {
                r := recover()
                c.So(r, ShouldNotBeNil)
            }()
            ValidateSubdomain(sd)
        }
        for _, sd := range badSubdomains {
            wraper(sd)
        }
    })
}

func Test_ValidateLanguageTypes(t *testing.T) {
    lts := []LanguageTypes{
        0,
        1,
    }
    Convey("Test ValidateLanguageTypes", t, func() {
        for _, l := range lts {
            ValidateLanguageTypes(l)
        }
    })
    
    wrong_lts := []LanguageTypes{
        8595,
        23000,
    }
    
    Convey("Test ValidateLanguageTypes (not found)", t, func(c C) {
        wraper := func(lt LanguageTypes) {
            defer func() {
                r := recover()
                c.So(r, ShouldNotBeNil)
            }()
            ValidateLanguageTypes(lt)
        }
        for _, lt := range wrong_lts {
            wraper(lt)
        }
    })
}

func Test_ValidateStrategyTypes(t *testing.T) {
    lts := []StrategyTypes{
        0,
        1,
        2,
    }
    Convey("Test ValidateStrategyTypes", t, func() {
        for _, l := range lts {
            ValidateStrategyTypes(l)
        }
    })
    wrong_lts := []StrategyTypes{
        8595,
        23000,
    }
    
    Convey("Test ValidateStrategyTypes (not found)", t, func(c C) {
        wraper := func(lt StrategyTypes) {
            defer func() {
                r := recover()
                c.So(r, ShouldNotBeNil)
            }()
            ValidateStrategyTypes(lt)
        }
        for _, lt := range wrong_lts {
            wraper(lt)
        }
    })
}

func Test_ValidateSecret(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Test_ValidateSecret", t, func() {
        Convey("should fails if service does not exists", func(c C) {
            defer func() {
                r := recover()
                if (r == nil) {
                    t.Error("did not panicked")
                    t.Fail()
                }
                switch x := r.(type) {
                case error:
                    c.So(merry.Is(x, NotFound), ShouldBeTrue)
                default:
                    t.Errorf("%v", x)
                    t.Fail()
                }
            }()
            CleanVault()
            user, err := NewUser("myroot")
            So(err, ShouldBeNil)
            user.TokenLogin()
            ValidateSecret("roles/admin/consul./consul/creds/readonly", user)
        })
        
        Convey("should fails if service does not contains this secret", func(c C) {
            defer func() {
                r := recover()
                if (r == nil) {
                    t.Error("did not panicked")
                    t.Fail()
                }
                switch x := r.(type) {
                case error:
                    c.So(merry.Is(x, NotFound), ShouldBeTrue)
                default:
                    t.Errorf("%v", x)
                    t.Fail()
                }
            }()
            CleanVault()
            PopulateVault()
            user, err := NewUser("myroot")
            So(err, ShouldBeNil)
            user.TokenLogin()
            ValidateSecret("roles/ml-team/consul./consul/creds/readonly", user)
        })
        
        Convey("should return vault path", func(c C) {
            CleanVault()
            PopulateVault()
            user, err := NewUser("myroot")
            So(err, ShouldBeNil)
            user.TokenLogin()
            key := ValidateSecret("roles/ml-team/gitlab./secret/gitlab/password", user)
            So(key, ShouldEqual, "secret/gitlab/password")
        })
        
    })
    
}
