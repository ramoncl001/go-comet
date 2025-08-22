package comet

// PoliciesConfig is a map with all controller policies
// configuration, such as Role, Permission or custom policies
type PoliciesConfig map[string][]Policy

type Policy struct {
	Validation AuthorizerFunction
	Value      interface{}
}

func Authorize(fn AuthorizerFunction, val interface{}) Policy {
	return Policy{
		Validation: fn,
		Value:      val,
	}
}

type AuthorizerFunction = func(RequestHandler, interface{}) RequestHandler

type AuthorizationMap map[interface{}]AuthorizerFunction
