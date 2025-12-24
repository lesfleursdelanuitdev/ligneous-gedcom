package types

// NameNode represents a structured name node.
// Wraps GedcomName to provide a node-like interface similar to elliotchance.
type NameNode struct {
	// Name is the parsed GedcomName
	Name *GedcomName

	// OriginalLine is the original GedcomLine for accessing sub-tags
	OriginalLine *GedcomLine
}

// NewNameNode creates a new NameNode from a GedcomName.
func NewNameNode(name *GedcomName) *NameNode {
	if name == nil {
		return nil
	}
	return &NameNode{Name: name}
}

// NewNameNodeFromLine creates a NameNode from a GedcomLine (NAME tag).
func NewNameNodeFromLine(line *GedcomLine) *NameNode {
	if line == nil || line.Tag != "NAME" {
		return nil
	}

	name, err := ParseName(line)
	if err != nil || name == nil {
		return nil
	}

	return &NameNode{
		Name:         name,
		OriginalLine: line,
	}
}

// GivenName returns the given name (first name).
func (nn *NameNode) GivenName() string {
	if nn == nil || nn.Name == nil {
		return ""
	}
	return nn.Name.Given
}

// Surname returns the surname (last name).
func (nn *NameNode) Surname() string {
	if nn == nil || nn.Name == nil {
		return ""
	}
	return nn.Name.Surname
}

// Prefix returns the name prefix (Dr., Mr., Mrs., etc.).
func (nn *NameNode) Prefix() string {
	if nn == nil || nn.Name == nil {
		return ""
	}
	return nn.Name.Prefix
}

// Suffix returns the name suffix (Jr., Sr., III, etc.).
func (nn *NameNode) Suffix() string {
	if nn == nil || nn.Name == nil {
		return ""
	}
	return nn.Name.Suffix
}

// Nickname returns the nickname.
func (nn *NameNode) Nickname() string {
	if nn == nil || nn.Name == nil {
		return ""
	}
	return nn.Name.Nickname
}

// SurnamePrefix returns the surname prefix (van, de, la, etc.).
func (nn *NameNode) SurnamePrefix() string {
	if nn == nil || nn.Name == nil {
		return ""
	}
	return nn.Name.SurnamePrefix
}

// Type returns the name type (birth, married, aka, etc.).
func (nn *NameNode) Type() NameType {
	if nn == nil || nn.Name == nil {
		return NameTypeUnknown
	}
	return nn.Name.Type
}

// String returns the formatted name string.
func (nn *NameNode) String() string {
	if nn == nil || nn.Name == nil {
		return ""
	}
	return nn.Name.Original
}

// IsValid returns true if the name is valid.
func (nn *NameNode) IsValid() bool {
	return nn != nil && nn.Name != nil && nn.Name.IsParsed
}

