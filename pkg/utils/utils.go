package menshend

import (
    "crypto/rand"
    "encoding/base64"
    "time"
    "strings"
    "github.com/ansel1/merry"
    "runtime"
    "github.com/Sirupsen/logrus"
)


//CheckPanic if error exist exit and log the file and line from where
//the errors comes
func CheckPanic(e error) {
    _, file, line, _ := runtime.Caller(1)
    if e != nil {
        logrus.WithFields(logrus.Fields{"file": file,
            "line": line,
        }).Panic(e)
    }
}
// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) []byte {
    b := make([]byte, n)
    _, err := rand.Read(b)
    // Note that err == nil only if we read len(b) bytes.
    CheckPanic(err)
    return b
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string) {
    b := GenerateRandomBytes(s)
    return base64.URLEncoding.EncodeToString(b)
}

func GetNow() int64 {
    return time.Now().Unix()
}

func getSubDomain(s string) string {
    domainParts := strings.Split(s, ".")
    if len(domainParts) < 3 {
        return ""
    }
    return domainParts[0]
}

func init() {
    merry.SetStackCaptureEnabled(false)
    
}

func SliceStringContains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func HttpCheckPanic(err error, userError merry.Error) {
    if err != nil {
        if (strings.Contains(err.Error(), "403")) {
            panic(PermissionError)
        }
        panic(userError.Append(err.Error()))
    }
}

func StringPtr(s string) *string {
    return &s
}
