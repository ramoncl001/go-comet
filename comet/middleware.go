package comet

// Middleware is a function that intercepts and processes HTTP requests
// before they reach the main handler, enabling cross-cutting concerns.
type Middleware = func(next RequestHandler) RequestHandler
