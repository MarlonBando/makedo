package engine

// RunContext holds the state for a single MakeDo execution session.
// This encapsulates the process registry and prepares the architecture for
// future context data like temporary filesystems or environment variables.
type RunContext struct {
	Registry *Registry
}

// NewRunContext initializes a new RunContext.
func NewRunContext() *RunContext {
	return &RunContext{
		Registry: NewRegistry(),
	}
}

// Cleanup performs teardown on the execution state, ensuring no zombie processes leak.
func (ctx *RunContext) Cleanup() {
	if ctx.Registry != nil {
		ctx.Registry.KillAll()
	}
}
