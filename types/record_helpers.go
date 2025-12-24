package types

// extractEvents extracts events from a record using the provided event tags.
// This is a shared helper for IndividualRecord.GetEvents() and FamilyRecord.GetEvents().
func extractEvents(record Record, eventTags []string) []map[string]interface{} {
	events := make([]map[string]interface{}, 0)
	for _, tag := range eventTags {
		eventLines := record.GetLines(tag)
		for _, line := range eventLines {
			event := map[string]interface{}{
				"type":        tag,
				"date":        line.GetValue("DATE"),
				"place":       line.GetValue("PLAC"),
				"description": line.Value,
			}
			events = append(events, event)
		}
	}
	return events
}

