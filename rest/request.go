package rest

type RequestMethod string

func (r RequestMethod) String() string {
	return string(r)
}

func (r RequestMethod) Method() string {
	switch r {
	case LIST:
		return "GET"
	default:
		return string(r)
	}
}

const (
	LIST   RequestMethod = "LIST"
	GET    RequestMethod = "GET"
	POST   RequestMethod = "POST"
	PUT    RequestMethod = "PUT"
	DELETE RequestMethod = "DELETE"
	PATCH  RequestMethod = "PATCH"
)
