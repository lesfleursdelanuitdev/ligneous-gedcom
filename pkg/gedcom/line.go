package gedcom

import "strings"

// GedcomLine represents a single line in a GEDCOM file with hierarchical structure.
type GedcomLine struct {
	Level      int                    // 0, 1, 2, etc.
	Tag        string                 // TAG name (e.g., "NAME", "BIRT")
	Value      string                 // Value after tag
	XrefID     string                 // Cross-reference ID (e.g., "@I1@")
	LineNumber int                    // Original line number in file
	Parent     *GedcomLine            // Parent line (nil for level 0)
	Children   map[string][]*GedcomLine // Children grouped by tag
}

// NewGedcomLine creates a new GedcomLine with the specified fields.
func NewGedcomLine(level int, tag, value, xrefID string) *GedcomLine {
	return &GedcomLine{
		Level:    level,
		Tag:      tag,
		Value:    value,
		XrefID:   xrefID,
		Children: make(map[string][]*GedcomLine),
	}
}

// AddChild adds a child line to this line and sets the child's parent.
func (gl *GedcomLine) AddChild(child *GedcomLine) {
	if gl.Children == nil {
		gl.Children = make(map[string][]*GedcomLine)
	}
	gl.Children[child.Tag] = append(gl.Children[child.Tag], child)
	child.Parent = gl
}

// GetValue retrieves a value using dot notation selector (e.g., "BIRT.DATE").
// Returns empty string if not found.
func (gl *GedcomLine) GetValue(selector string) string {
	if selector == "" {
		return gl.Value
	}

	parts := strings.Split(selector, ".")
	if len(parts) == 0 {
		return gl.Value
	}

	currentTag := parts[0]
	remaining := strings.Join(parts[1:], ".")

	if children, ok := gl.Children[currentTag]; ok {
		for _, child := range children {
			if len(parts) == 1 {
				return child.Value
			}
			if result := child.GetValue(remaining); result != "" {
				return result
			}
		}
	}

	return ""
}

// GetLines retrieves all lines matching the selector using dot notation.
func (gl *GedcomLine) GetLines(selector string) []*GedcomLine {
	if selector == "" {
		return []*GedcomLine{gl}
	}

	parts := strings.Split(selector, ".")
	currentTag := parts[0]
	remaining := strings.Join(parts[1:], ".")

	results := make([]*GedcomLine, 0)
	if children, ok := gl.Children[currentTag]; ok {
		for _, child := range children {
			if len(parts) == 1 {
				results = append(results, child)
			} else {
				results = append(results, child.GetLines(remaining)...)
			}
		}
	}

	return results
}

