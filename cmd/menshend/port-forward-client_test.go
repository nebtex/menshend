package main
/*
import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "time"
)

func TestClient(t *testing.T) {
    Convey("Test client handler", t, func() {
        Convey("Only the remote port is passed", func() {
            connect := func(v bool, k time.Duration, s string, r ...string) {
                So(s, ShouldEqual, "https://mongo.lab.example.com")
                So(r[0], ShouldEqual, "default.menshend:27017")
            }
            client([]string{"https://mongo.lab.example.com", "27017"}, connect)
        })
    
        Convey("local port and remote port are passed", func() {
            connect := func(v bool, k time.Duration, s string, r ...string) {
                So(s, ShouldEqual, "https://mongo.lab.example.com")
                So(r[0], ShouldEqual, "3000:default.menshend:27017")
            }
            client([]string{"https://mongo.lab.example.com", "3000:27017"}, connect)
        })
    
        Convey("local host and remote port are passed", func() {
            connect := func(v bool, k time.Duration, s string, r ...string) {
                So(s, ShouldEqual, "https://mongo.lab.example.com")
                So(r[0], ShouldEqual, "127.0.0.1:default.menshend:27017")
            }
            client([]string{"https://mongo.lab.example.com", "127.0.0.1:27017"}, connect)
        })
        Convey("local host local port and remote port are passed", func() {
            connect := func(v bool, k time.Duration, s string, r ...string) {
                So(v, ShouldEqual, true)
                So(s, ShouldEqual, "https://mongo.lab.example.com")
                So(r[0], ShouldEqual, "127.0.0.1:3000:default.menshend:27017")
            }
            client([]string{"--v", "https://mongo.lab.example.com", "127.0.0.1:3000:27017"}, connect)
        })
    })
    
}
*/
