package engine

// RunContext holds the state for a single MakeDo execution session.
type RunContext struct {
	Registry *Registry
}

func NewRunContext() *RunContext {
	return &RunContext{
		Registry: NewRegistry(),
	}
}

func (ctx *RunContext) Cleanup() {
	if ctx.Registry != nil {
		ctx.Registry.KillAll()
	}
}
