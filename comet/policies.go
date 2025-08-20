package comet

import (
	"github.com/ramoncl001/comet/ioc"
)

var RequireRole = func(next RequestHandler, value interface{}) RequestHandler {
	return func(req *Request) Response {
		sessionManager, err := ioc.ResolveSingleton[SessionManager](req.Context())
		if err != nil {
			return Error("could not resolve session manager")
		}

		claims, err := sessionManager.Validate(req)
		if err != nil {
			return Unauthorized()
		}

		if claims["role"] != value {
			return Unauthorized()
		}

		return next(req)
	}
}
