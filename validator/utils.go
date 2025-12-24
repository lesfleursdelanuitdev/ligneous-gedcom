package validator

import "regexp"

var (
	// xrefPattern matches valid xref format: @[A-Za-z0-9]{1,20}@
	xrefPattern = regexp.MustCompile(`^@[A-Za-z0-9]{1,20}@$`)
)

// isValidXref checks if an xref has valid format.
func isValidXref(xref string) bool {
	if len(xref) < 3 {
		return false
	}
	return xref[0] == '@' && xref[len(xref)-1] == '@'
}

// isValidXrefID checks if an xref ID has valid format using regex.
func isValidXrefID(xrefID string) bool {
	return xrefPattern.MatchString(xrefID)
}

// isUserDefinedTag checks if a tag is a user-defined tag (starts with underscore).
func isUserDefinedTag(tag string) bool {
	return len(tag) > 0 && tag[0] == '_'
}
