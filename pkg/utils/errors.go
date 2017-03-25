package menshend

import "github.com/ansel1/merry"
import "net/http"
//BadRequest ...
var InternalError = merry.New("Internal Error").WithHTTPCode(500).WithUserMessage("Internal Error")
var NotFound = merry.New("Resource not found").WithHTTPCode(404).WithUserMessage("Resource not found")
var NotAuthorized = merry.New("Not Authorized").WithHTTPCode(401).WithUserMessage("Not Authorized")
//PermissionError this mean that the acl token has not access to x key on consul
var PermissionError = merry.New("Permission Error").WithHTTPCode(403).WithUserMessage("Permission Error")
var BadRequest = merry.New("Bad request").WithHTTPCode(400).WithUserMessage("Bad request")
var BadGateway = merry.New("Bad Gateway").WithHTTPCode(http.StatusBadGateway).WithUserMessage("Bad Gateway")



