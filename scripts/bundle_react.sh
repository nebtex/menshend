#!/usr/bin/env bash
set -e
go get github.com/rakyll/statik
#pack static files
rm -rf statik
statik -src=ui

echo "package statik

import (
    \"net/http\"
    \"fmt\"
    \"github.com/gorilla/csrf\"
)


func Index() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        w.Write([]byte(fmt.Sprintf(\``cat ui/index.html`\`,  csrf.Token(r))))
    })
}" > statik/index.go
