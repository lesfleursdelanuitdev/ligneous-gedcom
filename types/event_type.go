package types

// EventType represents the type of a GEDCOM event.
// Supports all standard GEDCOM event types plus custom events.
type EventType string

// Standard GEDCOM event types (Individual events)
const (
	EventTypeBirth           EventType = "BIRT" // Birth
	EventTypeDeath           EventType = "DEAT" // Death
	EventTypeBurial          EventType = "BURI" // Burial
	EventTypeCremation       EventType = "CREM" // Cremation
	EventTypeChristening     EventType = "CHR"  // Christening
	EventTypeBaptism         EventType = "BAPM" // Baptism
	EventTypeBarMitzvah      EventType = "BARM" // Bar Mitzvah
	EventTypeBatMitzvah      EventType = "BASM" // Bat Mitzvah
	EventTypeBlessing        EventType = "BLES" // Blessing
	EventTypeAdultChristening EventType = "CHRA" // Adult Christening
	EventTypeConfirmation    EventType = "CONF" // Confirmation
	EventTypeFirstCommunion  EventType = "FCOM" // First Communion
	EventTypeOrdination      EventType = "ORDN" // Ordination
	EventTypeNaturalization  EventType = "NATU" // Naturalization
	EventTypeEmigration      EventType = "EMIG" // Emigration
	EventTypeImmigration     EventType = "IMMI" // Immigration
	EventTypeCensus          EventType = "CENS" // Census
	EventTypeProbate         EventType = "PROB" // Probate
	EventTypeWill            EventType = "WILL" // Will
	EventTypeGraduation      EventType = "GRAD" // Graduation
	EventTypeRetirement      EventType = "RETI" // Retirement
	EventTypeResidence       EventType = "RESI" // Residence
	EventTypeOccupation      EventType = "OCCU" // Occupation
	EventTypeEducation       EventType = "EDUC" // Education
	EventTypeCustom          EventType = "EVEN" // Custom/Generic event (requires TYPE sub-tag)
)

// Family event types
const (
	EventTypeMarriage        EventType = "MARR" // Marriage
	EventTypeDivorce         EventType = "DIV"  // Divorce
	EventTypeAnnulment       EventType = "ANUL" // Annulment
	EventTypeMarriageBann   EventType = "MARB" // Marriage Bann
	EventTypeMarriageContract EventType = "MARC" // Marriage Contract
	EventTypeMarriageLicense EventType = "MARL" // Marriage License
	EventTypeMarriageSettlement EventType = "MARS" // Marriage Settlement
	EventTypeEngagement      EventType = "ENGA" // Engagement
	EventTypeMarriageNotice  EventType = "MARB" // Marriage Notice
)

// Attribute types (also treated as events)
const (
	EventTypeCaste          EventType = "CAST" // Caste
	EventTypeDescription    EventType = "DSCR" // Physical Description
	EventTypeNationality    EventType = "NATI" // Nationality
	EventTypeProperty       EventType = "PROP" // Property
	EventTypeReligion       EventType = "RELI" // Religion
	EventTypeTitle          EventType = "TITL" // Title
)

// IsCustom returns true if this is a custom event type (EVEN with TYPE sub-tag).
func (et EventType) IsCustom() bool {
	return et == EventTypeCustom
}

// String returns the string representation of the event type.
func (et EventType) String() string {
	return string(et)
}

// IsValid returns true if this is a known standard event type.
func (et EventType) IsValid() bool {
	standardTypes := []EventType{
		EventTypeBirth, EventTypeDeath, EventTypeBurial, EventTypeCremation,
		EventTypeChristening, EventTypeBaptism, EventTypeBarMitzvah, EventTypeBatMitzvah,
		EventTypeBlessing, EventTypeAdultChristening, EventTypeConfirmation,
		EventTypeFirstCommunion, EventTypeOrdination, EventTypeNaturalization,
		EventTypeEmigration, EventTypeImmigration, EventTypeCensus, EventTypeProbate,
		EventTypeWill, EventTypeGraduation, EventTypeRetirement, EventTypeResidence,
		EventTypeOccupation, EventTypeEducation, EventTypeCustom,
		EventTypeMarriage, EventTypeDivorce, EventTypeAnnulment, EventTypeMarriageBann,
		EventTypeMarriageContract, EventTypeMarriageLicense, EventTypeMarriageSettlement,
		EventTypeEngagement, EventTypeMarriageNotice,
		EventTypeCaste, EventTypeDescription, EventTypeNationality, EventTypeProperty,
		EventTypeReligion, EventTypeTitle,
	}

	for _, st := range standardTypes {
		if et == st {
			return true
		}
	}

	return false
}

// ParseEventType parses a string into an EventType.
// For custom events (EVEN), the actual type comes from the TYPE sub-tag.
// This function handles the tag name, not the custom type value.
func ParseEventType(tag string) EventType {
	return EventType(tag)
}

