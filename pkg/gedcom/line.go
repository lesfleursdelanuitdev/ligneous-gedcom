package gedcom

import (
	"fmt"
	"sort"
	"strings"
)

// GedcomLine represents a single line in a GEDCOM file with hierarchical structure.
// Each line has a level (0, 1, 2, etc.), a tag (e.g., "NAME", "BIRT"), an optional
// value, and can have child lines that are subordinate to it.
//
// The hierarchical structure is maintained through:
//   - Parent pointer: Links to the parent line
//   - Children map: Groups child lines by tag for efficient access
//
// This structure allows representing the complete nested hierarchy of a GEDCOM file.
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

// ToGED converts the line and all its children to GEDCOM format.
// Returns a slice of strings, one per line.
// Children are sorted by tag for consistent output.
func (gl *GedcomLine) ToGED() []string {
	lines := []string{gl.toGEDLine()}
	
	// Collect all tags and sort them for consistent output
	tags := make([]string, 0, len(gl.Children))
	for tag := range gl.Children {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	
	// Process children in sorted tag order
	for _, tag := range tags {
		children := gl.Children[tag]
		for _, child := range children {
			lines = append(lines, child.ToGED()...)
		}
	}
	
	return lines
}

// toGEDLine converts a single line to GEDCOM format string.
func (gl *GedcomLine) toGEDLine() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("%d", gl.Level))
	
	if gl.XrefID != "" {
		parts = append(parts, gl.XrefID)
	}
	
	parts = append(parts, gl.Tag)
	
	if gl.Value != "" {
		parts = append(parts, gl.Value)
	}
	
	return strings.Join(parts, " ")
}

// SetValue sets a value using dot notation selector (e.g., "GEDC.VERS").
// Creates the path if it doesn't exist.
func (gl *GedcomLine) SetValue(selector string, value string) {
	if selector == "" {
		gl.Value = value
		return
	}

	parts := strings.Split(selector, ".")
	if len(parts) == 0 {
		gl.Value = value
		return
	}

	current := gl
	// Navigate/create path for all parts except the last
	for i := 0; i < len(parts)-1; i++ {
		tag := parts[i]
		if current.Children == nil {
			current.Children = make(map[string][]*GedcomLine)
		}
		
		children, exists := current.Children[tag]
		if !exists || len(children) == 0 {
			// Create new child line
			child := NewGedcomLine(current.Level+1, tag, "", "")
			current.AddChild(child)
			current = child
		} else {
			// Use first existing child
			current = children[0]
		}
	}

	// Set value on the last part
	lastTag := parts[len(parts)-1]
	if current.Children == nil {
		current.Children = make(map[string][]*GedcomLine)
	}
	
	children, exists := current.Children[lastTag]
	if !exists || len(children) == 0 {
		// Create new child line with value
		child := NewGedcomLine(current.Level+1, lastTag, value, "")
		current.AddChild(child)
	} else {
		// Update existing child's value
		children[0].Value = value
	}
}

