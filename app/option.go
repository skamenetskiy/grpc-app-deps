package app

type Option func(a *app)

func WithHTTPPort(port int) Option {
	return func(a *app) {
		a.httpPort = port
	}
}

func WithGRPCPort(port int) Option {
	return func(a *app) {
		a.grpcPort = port
	}
}

func WithSwaggerFile(data []byte) Option {
	return func(a *app) {
		a.swaggerFile = data
	}
}
