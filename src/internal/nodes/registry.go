package nodes

// Registry holds valid directive keywords.
// It uses a map for O(1) lookups.
type Registry struct {
	keywords map[string]struct{}
}

// NewRegistry creates an empty directive registry.
func NewRegistry() *Registry {
	return &Registry{
		keywords: make(map[string]struct{}),
	}
}

// Register adds a keyword to the registry.
func (r *Registry) Register(keyword string) {
	r.keywords[keyword] = struct{}{}
}

// IsValid checks if a keyword is registered.
func (r *Registry) IsValid(keyword string) bool {
	_, ok := r.keywords[keyword]
	return ok
}

// Keywords returns all registered keywords.
func (r *Registry) Keywords() []string {
	result := make([]string, 0, len(r.keywords))
	for k := range r.keywords {
		result = append(result, k)
	}
	return result
}
