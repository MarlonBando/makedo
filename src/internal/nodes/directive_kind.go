package nodes

type DirectiveKind uint8

const (
	DirectiveUnknown DirectiveKind = iota
	DirectiveOut
	DirectiveOutRegex
	DirectiveCmd
	DirectivePwd
	DirectiveCheckpath
	DirectiveSkip
	DirectiveExit
)

// For fewer keywords probably a slice would be faster
// but in the future prolly a map it's a better choice
// and doesn't make any difference to the user so
// map makes code cleaner
var directiveKeywords = map[string]DirectiveKind{
	"out":       DirectiveOut,
	"outr":      DirectiveOutRegex,
	"cmd":       DirectiveCmd,
	"pwd":       DirectivePwd,
	"checkpath": DirectiveCheckpath,
	"skip":      DirectiveSkip,
	"exit":      DirectiveExit,
}

// ParseDirectiveKind returns the DirectiveKind for a keyword.
// Returns DirectiveUnknown if the keyword is not recognized.
// NOTE: We allocate a string for lookup but it happens only
// when scanning the ast for makdedo block so it's fine
func ParseDirectiveKind(keyword []byte) DirectiveKind {
	kind, ok := directiveKeywords[string(keyword)]
	if !ok {
		return DirectiveUnknown
	}
	return kind
}

// IsValidKeyword checks if a keyword is a recognized directive.
func IsValidKeyword(keyword []byte) bool {
	return ParseDirectiveKind(keyword) != DirectiveUnknown
}

// Used only for unit testing
func (k DirectiveKind) String() string {
	switch k {
	case DirectiveOut:
		return "out"
	case DirectiveOutRegex:
		return "outr"
	case DirectiveCmd:
		return "cmd"
	case DirectivePwd:
		return "pwd"
	case DirectiveCheckpath:
		return "checkpath"
	case DirectiveSkip:
		return "skip"
	case DirectiveExit:
		return "exit"
	default:
		return "unknown"
	}
}
