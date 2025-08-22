package comet

type requestMethod string

func (r requestMethod) string() string {
	return string(r)
}

func (r requestMethod) method() string {
	switch r {
	case list:
		return "GET"
	default:
		return string(r)
	}
}

const (
	list   requestMethod = "LIST"
	get    requestMethod = "GET"
	post   requestMethod = "POST"
	put    requestMethod = "PUT"
	delete requestMethod = "DELETE"
	patch  requestMethod = "PATCH"
)
