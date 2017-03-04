package menshend

import "github.com/ansel1/merry"

//BadRequest ...
var InternalError = merry.New("Internal Error").WithHTTPCode(500)
var NotFound = merry.New("Resource not found").WithHTTPCode(404)
var NotAuthorized = merry.New("Not Authorized").WithHTTPCode(401)
//PermissionError this mean that the acl token has not access to x key on consul
var PermissionError = merry.New("Permission Error").WithHTTPCode(403)
var BadRequest = merry.New("Bad request").WithHTTPCode(400)

