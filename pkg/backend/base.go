package backend

type Backend interface {
    BaseUrl() string
    Headers() map[string]string
}
