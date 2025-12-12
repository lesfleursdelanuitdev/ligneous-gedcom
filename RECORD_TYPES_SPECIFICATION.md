# GEDCOM 5.5.1 Record Types Specification - Detailed Analysis

## Overview

This document provides comprehensive specifications for the key record types in GEDCOM 5.5.1: Individual (INDI), Family (FAM), Note (NOTE), and Events. This is planning documentation only - no code implementation.

---

## 1. INDIVIDUAL RECORD (INDI)

### Purpose
Represents a single person in the genealogical dataset, containing personal details, life events, relationships, and attributes.

### Structure
```
0 @XREF@ INDI
1 NAME <PERSONAL_NAME>
1 SEX <SEX_VALUE>
1 [BIRT | CHR | DEAT | BURI | CREM | ...] [<Y|<NULL>]
1 [ADOP | BAPM | BARM | BASM | BLES | CHRA | CONF | FCOM | ...]
1 [RESI | OCCU | EDUC | PROP | TITL | ...]
1 FAMS @XREF:FAM@
1 FAMC @XREF:FAM@
1 ASSO @XREF:INDI@
1 [NOTE | SOUR | OBJE | ...]
```

### Required Fields
- **XREF_ID**: Must be present (format: `@I1@`, `@I2@`, etc.)
- **NAME**: At least one NAME tag is typically required (spec may vary)

### Key Subfields

#### 1.1 NAME (Personal Name)
**Structure:**
```
1 NAME <PERSONAL_NAME> [ /<SURNAME>/ ]
2 TYPE <NAME_TYPE>
2 NPFX <NAME_PIECE_PREFIX>
2 GIVN <GIVEN_NAME>
2 NICK <NICKNAME>
2 SPFX <SURNAME_PREFIX>
2 SURN <SURNAME>
2 NSFX <NAME_PIECE_SUFFIX>
2 FONE <PHONETIC_NAME>
2 ROMN <ROMANIZED_NAME>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
```

**Key Points:**
- Multiple NAME tags allowed (different name variations)
- Surname enclosed in slashes: `John /Doe/`
- Components can be specified separately (GIVN, SURN, etc.)
- Supports phonetic (FONE) and romanized (ROMN) variations
- Can have source citations and notes

**Examples:**
```
1 NAME Robert Eugene /Williams/
2 SURN Williams
2 GIVN Robert Eugene

1 NAME Mary Ann /Wilson/
2 SURN Wilson
2 GIVN Mary Ann
```

#### 1.2 SEX (Gender)
**Values:**
- `M` - Male
- `F` - Female
- `U` - Unknown/Unspecified
- `X` - Intersex (in some implementations)
- `N` - Not applicable

#### 1.3 Family Relationships

**FAMS (Family as Spouse):**
- Links individual to family where they are a spouse
- Multiple FAMS allowed (multiple marriages)
- Format: `1 FAMS @F1@`

**FAMC (Family as Child):**
- Links individual to family where they are a child
- Multiple FAMC allowed (adoption, step-parents, etc.)
- Can have PEDI (Pedigree) qualifier:
  - `birth` - Biological
  - `adopted` - Adopted
  - `foster` - Foster
  - `sealing` - LDS sealing
- Format: 
```
1 FAMC @F1@
2 PEDI adopted
```

#### 1.4 Life Events

**Birth Event (BIRT):**
```
1 BIRT [<Y|<NULL>]
2 DATE <DATE_VALUE>
2 PLAC <PLACE_NAME>
2 ADDR <ADDRESS_STRUCTURE>
2 AGE <AGE_AT_EVENT>
2 CAUS <CAUSE_OF_EVENT>
2 AGNC <RESPONSIBLE_AGENCY>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
2 OBJE @XREF:OBJE@
```

**Death Event (DEAT):**
```
1 DEAT [<Y|<NULL>]
2 DATE <DATE_VALUE>
2 PLAC <PLACE_NAME>
2 ADDR <ADDRESS_STRUCTURE>
2 CAUS <CAUSE_OF_EVENT>
2 AGE <AGE_AT_EVENT>
2 AGNC <RESPONSIBLE_AGENCY>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
2 OBJE @XREF:OBJE@
```

**Other Individual Events:**
- `CHR` - Christening
- `BURI` - Burial
- `CREM` - Cremation
- `ADOP` - Adoption
- `BAPM` - Baptism
- `BARM` - Bar Mitzvah
- `BASM` - Bas Mitzvah
- `BLES` - Blessing
- `CHRA` - Adult Christening
- `CONF` - Confirmation
- `FCOM` - First Communion
- `ORDN` - Ordination
- `NATU` - Naturalization
- `EMIG` - Emigration
- `IMMI` - Immigration
- `CENS` - Census
- `PROB` - Probate
- `WILL` - Will
- `GRAD` - Graduation
- `RETI` - Retirement

**Event Structure:**
All events follow similar structure:
- DATE (optional)
- PLAC (optional)
- ADDR (optional)
- SOUR (optional, multiple)
- NOTE (optional, multiple)
- OBJE (optional, multiple)
- Event-specific tags (CAUS, AGE, AGNC, etc.)

#### 1.5 Attributes

**Residence (RESI):**
```
1 RESI [<RESIDENCE_DESCRIPTION>]
2 DATE <DATE_VALUE>
2 PLAC <PLACE_NAME>
2 ADDR <ADDRESS_STRUCTURE>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
```

**Occupation (OCCU):**
```
1 OCCU <OCCUPATION_DESCRIPTION>
2 DATE <DATE_VALUE>
2 PLAC <PLACE_NAME>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
```

**Other Attributes:**
- `CAST` - Caste
- `DSCR` - Physical Description
- `EDUC` - Education
- `NATI` - Nationality
- `NCHI` - Number of Children
- `NMR` - Number of Marriages
- `PROP` - Property/Posessions
- `RELI` - Religious Affiliation
- `TITL` - Title (Nobility, Rank, etc.)
- `FACT` - Fact (user-defined)

**Attribute Structure:**
- Value (optional)
- DATE (optional)
- PLAC (optional)
- SOUR (optional)
- NOTE (optional)

#### 1.6 Associations (ASSO)
Links to other individuals with relationship description:
```
1 ASSO @XREF:INDI@
2 RELA <RELATIONSHIP_DESCRIPTION>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
```

**Common RELA values:**
- `Godparent`
- `Witness`
- `Friend`
- `Colleague`
- Custom descriptions

#### 1.7 Other Tags
- `ALIA` - Alias (pointer to another INDI)
- `ANCI` - Ancestors interest
- `DESI` - Descendants interest
- `RFN` - Record file number
- `AFN` - Ancestral file number
- `REFN` - User reference number
- `RIN` - Record ID number
- `CHAN` - Change date
- `NOTE` - Notes (inline or pointer)
- `SOUR` - Source citations
- `OBJE` - Multimedia objects
- `SUBM` - Submitter

### Validation Rules
1. Must have XREF_ID
2. Should have at least one NAME
3. SEX should be M, F, or U
4. FAMS and FAMC must reference valid FAM records
5. Events should have DATE or PLAC (at least one)
6. Multiple events of same type allowed (e.g., multiple marriages via FAMS)

### Special Cases
- **Multiple Names**: Individual can have multiple NAME tags (maiden name, married name, etc.)
- **Multiple Marriages**: Multiple FAMS tags link to different families
- **Adoption**: FAMC with PEDI=adopted
- **Unknown Parents**: FAMC may be missing
- **Unmarried**: FAMS may be missing

---

## 2. FAMILY RECORD (FAM)

### Purpose
Represents a family unit, linking individuals as spouses and children, and documenting family events.

### Structure
```
0 @XREF@ FAM
1 HUSB @XREF:INDI@
1 WIFE @XREF:INDI@
1 CHIL @XREF:INDI@
1 [MARR | DIV | ANUL | ...] [<Y|<NULL>]
1 [CENS | EVEN | ...]
1 [NOTE | SOUR | OBJE | ...]
```

### Required Fields
- **XREF_ID**: Must be present (format: `@F1@`, `@F2@`, etc.)
- **At least one**: HUSB, WIFE, or CHIL (family must have at least one member)

### Key Subfields

#### 2.1 Spouse References

**HUSB (Husband):**
```
1 HUSB @XREF:INDI@
```
- Points to INDI record
- Optional (single-parent families)
- Only one HUSB per family

**WIFE (Wife):**
```
1 WIFE @XREF:INDI@
```
- Points to INDI record
- Optional (single-parent families)
- Only one WIFE per family

**Special Cases:**
- Single parent: Only HUSB or only WIFE
- Same-sex couples: May use HUSB for both or custom tags
- Multiple marriages: Individual has multiple FAMS, each pointing to different FAM records

#### 2.2 Children References

**CHIL (Child):**
```
1 CHIL @XREF:INDI@
```
- Points to INDI record
- Multiple CHIL tags allowed (one per child)
- **Preferred order**: Chronological by birth
- Children should have FAMC pointing back to this family

**Child Relationships:**
Children can specify relationship via FAMC:
```
1 FAMC @F1@
2 PEDI adopted
```

#### 2.3 Family Events

**Marriage Event (MARR):**
```
1 MARR [<Y|<NULL>]
2 DATE <DATE_VALUE>
2 PLAC <PLACE_NAME>
2 ADDR <ADDRESS_STRUCTURE>
2 TYPE <MARRIAGE_TYPE>
2 AGNC <RESPONSIBLE_AGENCY>
2 STAT <MARRIAGE_STATUS>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
2 OBJE @XREF:OBJE@
```

**Marriage Types (TYPE):**
- `CIVIL` - Civil marriage
- `RELIGIOUS` - Religious ceremony
- `COMMON` - Common law
- `PARTNERS` - Domestic partnership
- Custom values

**Marriage Status (STAT):**
- `MARRIED` - Currently married
- `DIVORCED` - Divorced
- `ANNULLED` - Annulled
- `UNKNOWN` - Unknown status

**Divorce Event (DIV):**
```
1 DIV [<Y|<NULL>]
2 DATE <DATE_VALUE>
2 PLAC <PLACE_NAME>
2 ADDR <ADDRESS_STRUCTURE>
2 STAT <DIVORCE_STATUS>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@
2 OBJE @XREF:OBJE@
```

**Other Family Events:**
- `ANUL` - Annulment
- `CENS` - Census
- `DIVF` - Divorce filed
- `ENGA` - Engagement
- `MARB` - Marriage banns
- `MARC` - Marriage contract
- `MARL` - Marriage license
- `MARS` - Marriage settlement
- `EVEN` - Generic event

**Event Structure:**
All family events follow similar structure:
- DATE (optional)
- PLAC (optional)
- ADDR (optional)
- SOUR (optional, multiple)
- NOTE (optional, multiple)
- OBJE (optional, multiple)
- Event-specific tags (TYPE, STAT, AGNC, etc.)

#### 2.4 Other Tags
- `NCHI` - Number of children
- `SUBM` - Submitter
- `REFN` - User reference number
- `RIN` - Record ID number
- `CHAN` - Change date
- `NOTE` - Notes (inline or pointer)
- `SOUR` - Source citations
- `OBJE` - Multimedia objects

### Validation Rules
1. Must have XREF_ID
2. Must have at least one: HUSB, WIFE, or CHIL
3. HUSB and WIFE must reference valid INDI records
4. CHIL must reference valid INDI records
5. Children should have FAMC pointing back to this family
6. Spouses should have FAMS pointing to this family
7. Events should have DATE or PLAC (at least one)
8. Multiple children allowed (multiple CHIL tags)

### Special Cases
- **Single Parent**: Only HUSB or only WIFE present
- **No Children**: Family with only spouses
- **Multiple Marriages**: Individual appears in multiple FAM records via FAMS
- **Adoption**: Child's FAMC has PEDI=adopted
- **Step-families**: Child has multiple FAMC (biological and step-parent)
- **Divorced/Remarried**: Individual has multiple FAMS, each with different spouse

---

## 3. NOTE RECORD (NOTE)

### Purpose
Contains additional textual information, commentary, or explanations that can be linked to other records.

### Structure
```
0 @XREF@ NOTE <SUBMITTER_TEXT>
1 CONT <SUBMITTER_TEXT>
1 CONC <SUBMITTER_TEXT>
1 SOUR @XREF:SOUR@
1 REFN <USER_REFERENCE_NUMBER>
1 RIN <RECORD_ID_NUMBER>
1 CHAN <CHANGE_DATE>
```

### Two Forms of Notes

#### 3.1 Inline Notes
Notes embedded directly in other records:
```
1 NOTE This is an inline note.
2 CONT This is a continuation line.
2 CONC This concatenates to the previous line.
```

**Structure:**
- `NOTE` - First line of note text
- `CONT` - Continuation (adds newline)
- `CONC` - Concatenation (no newline, direct append)

#### 3.2 Linked Notes (Note Records)
Notes stored as separate records and referenced:
```
0 @N1@ NOTE This is a linked note.
1 CONT It can be referenced by other records.
1 SOUR @S1@

0 @I1@ INDI
1 NOTE @N1@
```

**Structure:**
- Has XREF_ID (e.g., `@N1@`)
- Can be referenced by multiple records
- Can have source citations
- Can have change dates

### Key Subfields

**NOTE (Submitter Text):**
- First line of note content
- Can be empty if using CONT/CONC
- Format: `0 @XREF@ NOTE <TEXT>` or `1 NOTE <TEXT>`

**CONT (Continuation):**
- Continues note on new line
- Adds newline character before text
- Can have multiple CONT lines
- Format: `1 CONT <TEXT>` or `2 CONT <TEXT>`

**CONC (Concatenation):**
- Concatenates to previous line
- No newline, direct text append
- Can have multiple CONC lines
- Format: `1 CONC <TEXT>` or `2 CONC <TEXT>`

**Other Tags:**
- `SOUR` - Source citations
- `REFN` - User reference number
- `RIN` - Record ID number
- `CHAN` - Change date

### Usage Patterns

**Pattern 1: Short Inline Note**
```
1 NOTE Born in England, immigrated to USA in 1850.
```

**Pattern 2: Multi-line Inline Note**
```
1 NOTE This is a longer note that spans
2 CONT multiple lines. Each CONT adds
2 CONT a newline before the text.
```

**Pattern 3: Concatenated Note**
```
1 NOTE This is a long sentence that
2 CONC continues on the same line without
2 CONC a break, creating one continuous
2 CONC line of text.
```

**Pattern 4: Linked Note**
```
0 @N1@ NOTE This note is stored separately
1 CONT and can be referenced by multiple records.
1 SOUR @S1@

0 @I1@ INDI
1 NAME John /Doe/
1 NOTE @N1@

0 @F1@ FAM
1 HUSB @I1@
1 NOTE @N1@
```

### Validation Rules
1. Linked notes must have XREF_ID
2. Inline notes cannot have XREF_ID
3. CONT and CONC must follow a NOTE or another CONT/CONC
4. CONT/CONC cannot be subordinate to another CONT/CONC at lower level
5. Notes can be empty (just pointer: `1 NOTE @N1@`)

### Special Cases
- **Empty Notes**: `1 NOTE @N1@` (just a pointer, no text)
- **Very Long Notes**: Use multiple CONT lines
- **Formatted Text**: Some implementations use special formatting in notes
- **Source Citations**: Notes can cite sources
- **Shared Notes**: One note record referenced by multiple other records

---

## 4. EVENTS

### Purpose
Events represent occurrences in an individual's or family's life. They are not separate records but are embedded within INDI and FAM records.

### Event Structure
All events follow a common structure:
```
1 <EVENT_TAG> [<Y|<NULL>]
2 DATE <DATE_VALUE>
2 PLAC <PLACE_NAME>
2 ADDR <ADDRESS_STRUCTURE>
2 AGE <AGE_AT_EVENT>
2 CAUS <CAUSE_OF_EVENT>
2 AGNC <RESPONSIBLE_AGENCY>
2 TYPE <EVENT_TYPE>
2 STAT <EVENT_STATUS>
2 SOUR @XREF:SOUR@
2 NOTE @XREF:NOTE@ | <SUBMITTER_TEXT>
2 OBJE @XREF:OBJE@
```

### Event Components

#### 4.1 Event Tag
The event type identifier (BIRT, DEAT, MARR, etc.)

#### 4.2 Event Flag
- `<Y>` - Event occurred (explicit yes)
- `<NULL>` - Event occurred (implicit, no flag)
- Absent - Event may or may not have occurred

#### 4.3 DATE
- Format: Various (exact, approximate, ranges)
- Examples: `2 Oct 1822`, `BEF 1828`, `FROM 1900 TO 1905`
- Can be empty

#### 4.4 PLAC (Place)
- Hierarchical format: `City, State, Country`
- Can be empty
- May have FORM (place format) subfield

#### 4.5 ADDR (Address Structure)
```
2 ADDR <ADDRESS_LINE>
3 ADR1 <ADDRESS_LINE_1>
3 ADR2 <ADDRESS_LINE_2>
3 ADR3 <ADDRESS_LINE_3>
3 CITY <CITY>
3 STAE <STATE>
3 POST <POSTAL_CODE>
3 CTRY <COUNTRY>
```

#### 4.6 AGE
- Age at time of event
- Format: `25y 3m 2d` or `25`
- Used for events like DEAT (age at death)

#### 4.7 CAUS (Cause)
- Cause of event (typically for DEAT)
- Free text

#### 4.8 AGNC (Agency)
- Responsible agency (church, court, etc.)
- Free text

#### 4.9 TYPE
- Event type qualifier
- Used for MARR (CIVIL, RELIGIOUS, etc.)
- Used for EVEN (generic event type)

#### 4.10 STAT (Status)
- Event status
- Used for MARR (MARRIED, DIVORCED, etc.)
- Used for DIV (divorce status)

#### 4.11 SOUR (Source Citations)
- Multiple sources allowed
- Can have PAGE, DATA, etc. subfields
```
2 SOUR @S1@
3 PAGE Sec. 2, p. 45
3 DATA
4 DATE FROM Jan 1820 TO DEC 1825
```

#### 4.12 NOTE
- Inline or linked notes
- Multiple notes allowed

#### 4.13 OBJE (Multimedia)
- Links to multimedia objects
- Multiple objects allowed

### Individual Events

**Birth (BIRT):**
- Required for most individuals
- Should have DATE and/or PLAC
- Can have SOUR, NOTE, OBJE

**Death (DEAT):**
- Optional (living individuals)
- Can have CAUS (cause of death)
- Can have AGE (age at death)
- Can have SOUR, NOTE, OBJE

**Other Individual Events:**
- `CHR` - Christening
- `BURI` - Burial
- `CREM` - Cremation
- `ADOP` - Adoption
- `BAPM` - Baptism
- `BARM` - Bar Mitzvah
- `BASM` - Bas Mitzvah
- `BLES` - Blessing
- `CHRA` - Adult Christening
- `CONF` - Confirmation
- `FCOM` - First Communion
- `ORDN` - Ordination
- `NATU` - Naturalization
- `EMIG` - Emigration
- `IMMI` - Immigration
- `CENS` - Census
- `PROB` - Probate
- `WILL` - Will
- `GRAD` - Graduation
- `RETI` - Retirement
- `RESI` - Residence (attribute, but event-like)
- `OCCU` - Occupation (attribute, but event-like)
- `EVEN` - Generic event (with TYPE)

### Family Events

**Marriage (MARR):**
- Primary family event
- Should have DATE and/or PLAC
- Can have TYPE (CIVIL, RELIGIOUS, etc.)
- Can have STAT (MARRIED, DIVORCED, etc.)
- Can have SOUR, NOTE, OBJE

**Divorce (DIV):**
- Optional
- Should have DATE and/or PLAC
- Can have STAT (divorce status)
- Can have SOUR, NOTE, OBJE

**Other Family Events:**
- `ANUL` - Annulment
- `CENS` - Census
- `DIVF` - Divorce filed
- `ENGA` - Engagement
- `MARB` - Marriage banns
- `MARC` - Marriage contract
- `MARL` - Marriage license
- `MARS` - Marriage settlement
- `EVEN` - Generic event (with TYPE)

### Event Validation Rules
1. Events should have at least DATE or PLAC (recommended, not required)
2. Multiple events of same type allowed (e.g., multiple marriages)
3. SOUR, NOTE, OBJE can appear multiple times
4. DATE format must be valid GEDCOM date
5. PLAC should follow hierarchical format
6. AGE should be valid format
7. Event flag (Y) is optional

### Special Cases
- **Multiple Events**: Individual can have multiple events of same type (e.g., multiple marriages via different FAMS)
- **Unknown Dates**: DATE can be approximate (ABT, BEF, AFT, etc.)
- **Date Ranges**: FROM...TO, BET...AND formats
- **Generic Events**: EVEN tag with TYPE qualifier for custom events
- **Event Attributes**: Some events are actually attributes (OCCU, RESI) but follow event structure

---

## 5. Relationships and Cross-References

### Individual to Family Links
- **FAMS**: Individual → Family (as spouse)
- **FAMC**: Individual → Family (as child)
- **FAM**: Family → Individual (HUSB, WIFE, CHIL)

### Validation Requirements
1. **Bidirectional Links**: 
   - If INDI has FAMS @F1@, then FAM @F1@ should have HUSB or WIFE pointing to that INDI
   - If INDI has FAMC @F1@, then FAM @F1@ should have CHIL pointing to that INDI
   - If FAM has HUSB @I1@, then INDI @I1@ should have FAMS pointing to that FAM

2. **XREF Validation**:
   - All XREF pointers must reference existing records
   - XREF format must be valid (@...@)
   - XREF must be unique within file

3. **Circular References**:
   - Individual cannot be their own parent/child
   - Family cannot reference non-existent individuals

---

## 6. Implementation Considerations

### Data Structures Needed

**For INDI Records:**
- Name structure (multiple names, components)
- Event collection (multiple events per type)
- Attribute collection
- Relationship links (FAMS, FAMC)
- Association links (ASSO)

**For FAM Records:**
- Spouse references (HUSB, WIFE)
- Child references (CHIL collection)
- Event collection (MARR, DIV, etc.)

**For NOTE Records:**
- Text content (with CONT/CONC handling)
- Source citations
- Change tracking

**For Events:**
- Date parsing (complex formats)
- Place parsing (hierarchical)
- Address structure
- Source citations
- Multimedia links

### Validation Requirements

1. **Structure Validation**:
   - Required fields present
   - Valid XREF formats
   - Valid tag combinations

2. **Relationship Validation**:
   - Bidirectional link consistency
   - No circular references
   - Valid XREF targets

3. **Data Validation**:
   - Date format validity
   - Place format validity
   - SEX value validity
   - Event flag validity

### Performance Considerations

1. **XREF Indexing**: Build index for fast lookup
2. **Relationship Traversal**: Efficient parent/child/spouse navigation
3. **Event Queries**: Fast event lookup by type
4. **Note Resolution**: Efficient inline vs. linked note handling

---

## 7. Summary

### INDI Records
- Represent individuals with names, events, attributes
- Link to families via FAMS (spouse) and FAMC (child)
- Support multiple names, events, and relationships
- Rich event and attribute structure

### FAM Records
- Represent family units with spouses and children
- Link to individuals via HUSB, WIFE, CHIL
- Support family events (marriage, divorce, etc.)
- Enable complex family structures (multiple marriages, adoptions)

### NOTE Records
- Provide additional textual information
- Can be inline or linked (separate records)
- Support multi-line text via CONT/CONC
- Can be shared across multiple records

### Events
- Embedded in INDI and FAM records
- Common structure: DATE, PLAC, SOUR, NOTE, OBJE
- Support complex date formats
- Support source citations and multimedia
- Enable rich genealogical documentation

This comprehensive understanding is essential for designing accurate data structures and parsing algorithms in the Go implementation.

