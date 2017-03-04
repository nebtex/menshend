package menshend

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

//TestGetUser
func TestGetSubDomain(t *testing.T) {
	Convey("Should return subdomain name", t, func() {
		table := []struct {
			Domain    string
			SubDomain string
		}{{"criloz.nebtex.com",
			"criloz"},
			{"nebtex.com",
				""},
			{"xxx.genos.nebtex.com",
				"xxx"}}
		for _, test := range table {
			sd := getSubDomain(test.Domain)
			So(test.SubDomain, ShouldEqual, sd)
		}
		
	})
	
}
