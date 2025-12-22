package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// BuildGraphHybrid builds a graph using hybrid storage (SQLite + BadgerDB)
func BuildGraphHybrid(tree *gedcom.GedcomTree, sqlitePath, badgerPath string) (*Graph, error) {
	// Initialize hybrid storage
	storage, err := NewHybridStorage(sqlitePath, badgerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hybrid storage: %w", err)
	}

	// Create graph structure (will use hybrid storage)
	graph := NewGraph(tree)
	graph.hybridStorage = storage
	graph.hybridMode = true
	
	// Initialize query helpers with prepared statements
	queryHelpers, err := NewHybridQueryHelpers(storage.SQLite())
	if err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to create query helpers: %w", err)
	}
	graph.queryHelpers = queryHelpers
	
	// Initialize hybrid cache with default sizes
	// Node cache: 50K nodes (configurable)
	// XREF cache: 25K entries (configurable)
	// Query cache: 5K queries (configurable)
	hybridCache, err := NewHybridCache(50000, 25000, 5000)
	if err != nil {
		queryHelpers.Close()
		storage.Close()
		return nil, fmt.Errorf("failed to create hybrid cache: %w", err)
	}
	graph.hybridCache = hybridCache

	// Build graph in both databases
	if err := buildGraphInSQLite(storage, tree, graph); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to build SQLite indexes: %w", err)
	}

	if err := buildGraphInBadgerDB(storage, tree, graph); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to build BadgerDB graph: %w", err)
	}

	return graph, nil
}

// buildGraphInSQLite builds indexes in SQLite
func buildGraphInSQLite(storage *HybridStorage, tree *gedcom.GedcomTree, graph *Graph) error {
	db := storage.SQLite()

	// Start transaction for batch inserts
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statements for batch inserts
	stmtNode, err := tx.Prepare(`
		INSERT INTO nodes (id, xref, type, name, name_lower, birth_date, birth_place, sex, 
		                   has_children, has_spouse, living, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare node statement: %w", err)
	}
	defer stmtNode.Close()

	stmtXref, err := tx.Prepare(`
		INSERT INTO xref_mapping (xref, node_id) VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare xref statement: %w", err)
	}
	defer stmtXref.Close()

	// Process individuals
	individuals := tree.GetAllIndividuals()
	now := time.Now().Unix()

	for xrefID, record := range individuals {
		indiRecord, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		// Get or create node ID
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}

		// Extract indexed fields
		name := indiRecord.GetName()
		nameLower := toLower(name)
		birthDate := parseBirthDate(indiRecord)
		birthPlace := indiRecord.GetBirthPlace()
		sex := indiRecord.GetSex()
		
		// Determine boolean flags (will be updated after edges are processed)
		hasChildren := false // Will be updated later
		hasSpouse := false   // Will be updated later
		living := indiRecord.GetDeathDate() == ""

		// Insert into nodes table
		_, err := stmtNode.Exec(
			nodeID, xrefID, "individual", name, nameLower,
			birthDate, birthPlace, sex,
			boolToInt(hasChildren), boolToInt(hasSpouse), boolToInt(living),
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping
		_, err = stmtXref.Exec(xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert xref mapping %s: %w", xrefID, err)
		}
	}

	// Process families
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		_, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		// Get or create node ID
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}

		// Families don't have as many indexed fields
		_, err := stmtNode.Exec(
			nodeID, xrefID, "family", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert family node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping
		_, err = stmtXref.Exec(xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert family xref mapping %s: %w", xrefID, err)
		}
	}

	// Process notes
	notes := tree.GetAllNotes()
	for xrefID, record := range notes {
		_, ok := record.(*gedcom.NoteRecord)
		if !ok {
			continue
		}

		// Get or create node ID
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}

		// Notes don't have indexed fields in SQLite (for now)
		_, err := stmtNode.Exec(
			nodeID, xrefID, "note", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert note node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping
		_, err = stmtXref.Exec(xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert note xref mapping %s: %w", xrefID, err)
		}
	}

	// Process sources
	sources := tree.GetAllSources()
	for xrefID, record := range sources {
		_, ok := record.(*gedcom.SourceRecord)
		if !ok {
			continue
		}

		// Get or create node ID
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}

		// Sources don't have indexed fields in SQLite (for now)
		_, err := stmtNode.Exec(
			nodeID, xrefID, "source", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert source node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping
		_, err = stmtXref.Exec(xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert source xref mapping %s: %w", xrefID, err)
		}
	}

	// Process repositories
	repositories := tree.GetAllRepositories()
	for xrefID, record := range repositories {
		_, ok := record.(*gedcom.RepositoryRecord)
		if !ok {
			continue
		}

		// Get or create node ID
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}

		// Repositories don't have indexed fields in SQLite (for now)
		_, err := stmtNode.Exec(
			nodeID, xrefID, "repository", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert repository node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping
		_, err = stmtXref.Exec(xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert repository xref mapping %s: %w", xrefID, err)
		}
	}

	// Process event nodes (from individuals and families)
	// Events are embedded in records, so we need to extract them
	individualsForEvents := tree.GetAllIndividuals()
	for xrefID, record := range individualsForEvents {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		events := indi.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)

			// Get or create node ID
			nodeID := graph.xrefToID[eventID]
			if nodeID == 0 {
				nodeID = graph.nextID
				graph.nextID++
				graph.xrefToID[eventID] = nodeID
				graph.idToXref[nodeID] = eventID
			}

			// Events don't have indexed fields in SQLite (for now)
			_, err := stmtNode.Exec(
				nodeID, eventID, "event", "", "",
				nil, "", "",
				0, 0, 0,
				now, now,
			)
			if err != nil {
				return fmt.Errorf("failed to insert event node %s: %w", eventID, err)
			}

			// Insert into xref_mapping
			_, err = stmtXref.Exec(eventID, nodeID)
			if err != nil {
				return fmt.Errorf("failed to insert event xref mapping %s: %w", eventID, err)
			}
		}
	}

	// Process family events
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		// Check for MARR, DIV events
		eventTypes := []string{"MARR", "DIV", "ANUL", "ENGA", "MARB", "MARC", "MARL", "MARS"}
		for _, eventType := range eventTypes {
			eventLines := fam.GetLines(eventType)
			for i := range eventLines {
				eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)

				// Get or create node ID
				nodeID := graph.xrefToID[eventID]
				if nodeID == 0 {
					nodeID = graph.nextID
					graph.nextID++
					graph.xrefToID[eventID] = nodeID
					graph.idToXref[nodeID] = eventID
				}

				// Events don't have indexed fields in SQLite (for now)
				_, err := stmtNode.Exec(
					nodeID, eventID, "event", "", "",
					nil, "", "",
					0, 0, 0,
					now, now,
				)
				if err != nil {
					return fmt.Errorf("failed to insert event node %s: %w", eventID, err)
				}

				// Insert into xref_mapping
				_, err = stmtXref.Exec(eventID, nodeID)
				if err != nil {
					return fmt.Errorf("failed to insert event xref mapping %s: %w", eventID, err)
				}
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// buildGraphInBadgerDB stores graph structure in BadgerDB
func buildGraphInBadgerDB(storage *HybridStorage, tree *gedcom.GedcomTree, graph *Graph) error {
	db := storage.BadgerDB()

	// Process individuals and store in BadgerDB
	individuals := tree.GetAllIndividuals()
	
	// Use batch write for better performance
	writeBatch := db.NewWriteBatch()
	defer writeBatch.Cancel()

	for xrefID, record := range individuals {
		indiRecord, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		// Get node ID
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			continue // Skip if not in mapping
		}

		// Create IndividualNode for serialization
		node := &IndividualNode{
			BaseNode: &BaseNode{
				xrefID:   xrefID,
				nodeType: NodeTypeIndividual,
				record:   indiRecord,
			},
			Individual: indiRecord,
		}

		// Serialize and store
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set node %s: %w", xrefID, err)
		}
	}

	// Process families
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		famRecord, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		// Get node ID
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			continue
		}

		// Create FamilyNode for serialization
		node := &FamilyNode{
			BaseNode: &BaseNode{
				xrefID:   xrefID,
				nodeType: NodeTypeFamily,
				record:   famRecord,
			},
			Family: famRecord,
		}

		// Serialize and store
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize family node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set family node %s: %w", xrefID, err)
		}
	}

	// Process notes
	notes := tree.GetAllNotes()
	for xrefID, record := range notes {
		noteRecord, ok := record.(*gedcom.NoteRecord)
		if !ok {
			continue
		}

		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			continue
		}

		node := NewNoteNode(xrefID, noteRecord)
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize note node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set note node %s: %w", xrefID, err)
		}
	}

	// Process sources
	sources := tree.GetAllSources()
	for xrefID, record := range sources {
		sourceRecord, ok := record.(*gedcom.SourceRecord)
		if !ok {
			continue
		}

		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			continue
		}

		node := NewSourceNode(xrefID, sourceRecord)
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize source node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set source node %s: %w", xrefID, err)
		}
	}

	// Process repositories
	repositories := tree.GetAllRepositories()
	for xrefID, record := range repositories {
		repoRecord, ok := record.(*gedcom.RepositoryRecord)
		if !ok {
			continue
		}

		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			continue
		}

		node := NewRepositoryNode(xrefID, repoRecord)
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize repository node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set repository node %s: %w", xrefID, err)
		}
	}

	// Process event nodes (from individuals)
	individualsForEvents := tree.GetAllIndividuals()
	for xrefID, record := range individualsForEvents {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		indiNode := graph.GetIndividual(xrefID)
		if indiNode == nil {
			continue
		}

		events := indi.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
			eventNodeID := graph.xrefToID[eventID]
			if eventNodeID == 0 {
				continue
			}

			eventNode := NewEventNode(eventID, eventType, eventData)
			eventNode.Owner = indiNode

			data, err := SerializeNode(eventNode, graph)
			if err != nil {
				return fmt.Errorf("failed to serialize event node %s: %w", eventID, err)
			}

			key := fmt.Sprintf("node:%d", eventNodeID)
			if err := writeBatch.Set([]byte(key), data); err != nil {
				return fmt.Errorf("failed to set event node %s: %w", eventID, err)
			}
		}
	}

	// Process family events
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		famNode := graph.GetFamily(xrefID)
		if famNode == nil {
			continue
		}

		eventTypes := []string{"MARR", "DIV", "ANUL", "ENGA", "MARB", "MARC", "MARL", "MARS"}
		for _, eventType := range eventTypes {
			eventLines := fam.GetLines(eventType)
			for i := range eventLines {
				eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
				eventNodeID := graph.xrefToID[eventID]
				if eventNodeID == 0 {
					continue
				}

				// Create event data from line
				eventLine := eventLines[i]
				eventData := map[string]interface{}{
					"type":        eventType,
					"date":        eventLine.GetValue("DATE"),
					"place":       eventLine.GetValue("PLAC"),
					"description": eventLine.Value,
				}

				eventNode := NewEventNode(eventID, eventType, eventData)
				eventNode.Owner = famNode

				data, err := SerializeNode(eventNode, graph)
				if err != nil {
					return fmt.Errorf("failed to serialize event node %s: %w", eventID, err)
				}

				key := fmt.Sprintf("node:%d", eventNodeID)
				if err := writeBatch.Set([]byte(key), data); err != nil {
					return fmt.Errorf("failed to set event node %s: %w", eventID, err)
				}
			}
		}
	}

	// Flush batch
	if err := writeBatch.Flush(); err != nil {
		return fmt.Errorf("failed to flush BadgerDB batch: %w", err)
	}

	// Now process edges (after all nodes are stored)
	if err := buildEdgesInBadgerDB(db, tree, graph); err != nil {
		return fmt.Errorf("failed to build edges: %w", err)
	}

	// Update SQLite indexes with relationship flags (has_children, has_spouse)
	// This needs to be done after edges are processed
	if err := updateRelationshipFlags(storage, tree, graph); err != nil {
		return fmt.Errorf("failed to update relationship flags: %w", err)
	}

	return nil
}

// updateRelationshipFlags updates has_children and has_spouse flags in SQLite
func updateRelationshipFlags(storage *HybridStorage, tree *gedcom.GedcomTree, graph *Graph) error {
	db := storage.SQLite()

	// Process families to determine relationships
	families := tree.GetAllFamilies()
	
	// Track which individuals have children/spouses
	hasChildren := make(map[uint32]bool)
	hasSpouse := make(map[uint32]bool)

	for _, record := range families {
		famRecord, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		// Get family node ID
		famXref := famRecord.XrefID()
		famNodeID := graph.xrefToID[famXref]
		if famNodeID == 0 {
			continue
		}

		// Check husband
		husbandXref := famRecord.GetHusband()
		if husbandXref != "" {
			husbandID := graph.xrefToID[husbandXref]
			if husbandID != 0 {
				hasSpouse[husbandID] = true
			}
		}

		// Check wife
		wifeXref := famRecord.GetWife()
		if wifeXref != "" {
			wifeID := graph.xrefToID[wifeXref]
			if wifeID != 0 {
				hasSpouse[wifeID] = true
			}
		}

		// Check children
		children := famRecord.GetChildren()
		for _, childXref := range children {
			childID := graph.xrefToID[childXref]
			if childID != 0 {
				// Child has parents (but we're tracking if parents have children)
				// So we need to mark the parents
				if husbandID := graph.xrefToID[husbandXref]; husbandID != 0 {
					hasChildren[husbandID] = true
				}
				if wifeID := graph.xrefToID[wifeXref]; wifeID != 0 {
					hasChildren[wifeID] = true
				}
			}
		}
	}

	// Update SQLite with relationship flags
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE nodes SET has_children = ?, has_spouse = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	// Update all individuals
	individuals := tree.GetAllIndividuals()
	for xrefID := range individuals {
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			continue
		}

		hasChildrenVal := boolToInt(hasChildren[nodeID])
		hasSpouseVal := boolToInt(hasSpouse[nodeID])

		_, err := stmt.Exec(hasChildrenVal, hasSpouseVal, nodeID)
		if err != nil {
			return fmt.Errorf("failed to update node %d: %w", nodeID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// buildEdgesInBadgerDB stores edges in BadgerDB
func buildEdgesInBadgerDB(db *badger.DB, tree *gedcom.GedcomTree, graph *Graph) error {
	// Process families to extract edges
	families := tree.GetAllFamilies()
	writeBatch := db.NewWriteBatch()
	defer writeBatch.Cancel()

	// Track edges per node for efficient storage
	nodeEdges := make(map[uint32][]EdgeData)

	for xrefID, record := range families {
		famRecord, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		famNodeID := graph.xrefToID[xrefID]
		if famNodeID == 0 {
			continue
		}

		// Get husband
		husbandXref := famRecord.GetHusband()
		if husbandXref != "" {
			husbandID := graph.xrefToID[husbandXref]
			if husbandID != 0 {
				// HUSB edge: Family -> Individual
				edgeData := EdgeData{
					FromID:    famNodeID,
					ToID:      husbandID,
					EdgeType:  EdgeTypeHUSB,
					FamilyID:  famNodeID,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)

				// FAMS edge: Individual -> Family (reverse)
				edgeData2 := EdgeData{
					FromID:    husbandID,
					ToID:      famNodeID,
					EdgeType:  EdgeTypeFAMS,
					FamilyID:  famNodeID,
					Direction: DirectionBackward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[husbandID] = append(nodeEdges[husbandID], edgeData2)
			}
		}

		// Get wife
		wifeXref := famRecord.GetWife()
		if wifeXref != "" {
			wifeID := graph.xrefToID[wifeXref]
			if wifeID != 0 {
				// WIFE edge: Family -> Individual
				edgeData := EdgeData{
					FromID:    famNodeID,
					ToID:      wifeID,
					EdgeType:  EdgeTypeWIFE,
					FamilyID:  famNodeID,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)

				// FAMS edge: Individual -> Family (reverse)
				edgeData2 := EdgeData{
					FromID:    wifeID,
					ToID:      famNodeID,
					EdgeType:  EdgeTypeFAMS,
					FamilyID:  famNodeID,
					Direction: DirectionBackward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[wifeID] = append(nodeEdges[wifeID], edgeData2)
			}
		}

		// Get children
		children := famRecord.GetChildren()
		for _, childXref := range children {
			childID := graph.xrefToID[childXref]
			if childID != 0 {
				// CHIL edge: Family -> Individual
				edgeData := EdgeData{
					FromID:    famNodeID,
					ToID:      childID,
					EdgeType:  EdgeTypeCHIL,
					FamilyID:  famNodeID,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)

				// FAMC edge: Individual -> Family (reverse)
				edgeData2 := EdgeData{
					FromID:    childID,
					ToID:      famNodeID,
					EdgeType:  EdgeTypeFAMC,
					FamilyID:  famNodeID,
					Direction: DirectionBackward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[childID] = append(nodeEdges[childID], edgeData2)
			}
		}
	}

	// Process individuals for NOTE, SOUR edges
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		indiNodeID := graph.xrefToID[xrefID]
		if indiNodeID == 0 {
			continue
		}

		// NOTE references
		notes := indi.GetNotes()
		for _, noteXref := range notes {
			noteNodeID := graph.xrefToID[noteXref]
			if noteNodeID != 0 {
				edgeData := EdgeData{
					FromID:    indiNodeID,
					ToID:      noteNodeID,
					EdgeType:  EdgeTypeNOTE,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[indiNodeID] = append(nodeEdges[indiNodeID], edgeData)
			}
		}

		// SOUR references
		sources := indi.GetSources()
		for _, sourceXref := range sources {
			sourceNodeID := graph.xrefToID[sourceXref]
			if sourceNodeID != 0 {
				edgeData := EdgeData{
					FromID:    indiNodeID,
					ToID:      sourceNodeID,
					EdgeType:  EdgeTypeSOUR,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[indiNodeID] = append(nodeEdges[indiNodeID], edgeData)
			}
		}
	}

	// Process families for NOTE, SOUR edges
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		famNodeID := graph.xrefToID[xrefID]
		if famNodeID == 0 {
			continue
		}

		// NOTE references
		notes := fam.GetNotes()
		for _, noteXref := range notes {
			noteNodeID := graph.xrefToID[noteXref]
			if noteNodeID != 0 {
				edgeData := EdgeData{
					FromID:    famNodeID,
					ToID:      noteNodeID,
					EdgeType:  EdgeTypeNOTE,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)
			}
		}

		// SOUR references
		sources := fam.GetSources()
		for _, sourceXref := range sources {
			sourceNodeID := graph.xrefToID[sourceXref]
			if sourceNodeID != 0 {
				edgeData := EdgeData{
					FromID:    famNodeID,
					ToID:      sourceNodeID,
					EdgeType:  EdgeTypeSOUR,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)
			}
		}
	}

	// Process notes for SOUR edges
	notes := tree.GetAllNotes()
	for xrefID, record := range notes {
		note, ok := record.(*gedcom.NoteRecord)
		if !ok {
			continue
		}

		noteNodeID := graph.xrefToID[xrefID]
		if noteNodeID == 0 {
			continue
		}

		sources := note.GetValues("SOUR")
		for _, sourceXref := range sources {
			sourceNodeID := graph.xrefToID[sourceXref]
			if sourceNodeID != 0 {
				edgeData := EdgeData{
					FromID:    noteNodeID,
					ToID:      sourceNodeID,
					EdgeType:  EdgeTypeSOUR,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[noteNodeID] = append(nodeEdges[noteNodeID], edgeData)
			}
		}
	}

	// Process sources for REPO edges
	sources := tree.GetAllSources()
	for xrefID, record := range sources {
		source, ok := record.(*gedcom.SourceRecord)
		if !ok {
			continue
		}

		sourceNodeID := graph.xrefToID[xrefID]
		if sourceNodeID == 0 {
			continue
		}

		repos := source.GetValues("REPO")
		for _, repoXref := range repos {
			repoNodeID := graph.xrefToID[repoXref]
			if repoNodeID != 0 {
				edgeData := EdgeData{
					FromID:    sourceNodeID,
					ToID:      repoNodeID,
					EdgeType:  EdgeTypeREPO,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[sourceNodeID] = append(nodeEdges[sourceNodeID], edgeData)
			}
		}
	}

	// Process individuals for has_event edges
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		indiNodeID := graph.xrefToID[xrefID]
		if indiNodeID == 0 {
			continue
		}

		events := indi.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
			eventNodeID := graph.xrefToID[eventID]
			if eventNodeID != 0 {
				edgeData := EdgeData{
					FromID:    indiNodeID,
					ToID:      eventNodeID,
					EdgeType:  EdgeTypeHasEvent,
					Direction: DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[indiNodeID] = append(nodeEdges[indiNodeID], edgeData)
			}
		}
	}

	// Process families for has_event edges (MARR, DIV, etc.)
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		famNodeID := graph.xrefToID[xrefID]
		if famNodeID == 0 {
			continue
		}

		// Check for MARR, DIV events
		eventTypes := []string{"MARR", "DIV", "ANUL", "ENGA", "MARB", "MARC", "MARL", "MARS"}
		for _, eventType := range eventTypes {
			eventLines := fam.GetLines(eventType)
			for i := range eventLines {
				eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
				eventNodeID := graph.xrefToID[eventID]
				if eventNodeID != 0 {
					edgeData := EdgeData{
						FromID:    famNodeID,
						ToID:      eventNodeID,
						EdgeType:  EdgeTypeHasEvent,
						Direction: DirectionForward,
						Properties: make(map[string]interface{}),
					}
					nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)
				}
			}
		}
	}

	// Store edges in BadgerDB (grouped by node for efficient retrieval)
	for nodeID, edges := range nodeEdges {
		// Serialize edges
		data, err := serializeEdgeDataList(edges)
		if err != nil {
			return fmt.Errorf("failed to serialize edges for node %d: %w", nodeID, err)
		}

		// Store outgoing edges
		key := fmt.Sprintf("edges:%d:out", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set edges for node %d: %w", nodeID, err)
		}
	}

	// Flush batch
	if err := writeBatch.Flush(); err != nil {
		return fmt.Errorf("failed to flush edges batch: %w", err)
	}

	return nil
}

// Helper functions

func toLower(s string) string {
	// Simple lowercase conversion
	// In production, might want to use proper Unicode handling
	return strings.ToLower(s)
}

func parseBirthDate(indi *gedcom.IndividualRecord) *int64 {
	birthDateStr := indi.GetBirthDate()
	if birthDateStr == "" {
		return nil
	}

	// Try to parse the date
	// This is simplified - in production, use proper date parsing
	// For now, return nil (will be improved)
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

