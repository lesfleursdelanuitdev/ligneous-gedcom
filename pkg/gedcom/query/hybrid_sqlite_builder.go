package query

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

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

	now := time.Now().Unix()

	// Process all record types
	if err := processIndividualsForSQLite(tree, graph, stmtNode, stmtXref, now); err != nil {
		return err
	}

	if err := processFamiliesForSQLite(tree, graph, stmtNode, stmtXref, now); err != nil {
		return err
	}

	if err := processNotesForSQLite(tree, graph, stmtNode, stmtXref, now); err != nil {
		return err
	}

	if err := processSourcesForSQLite(tree, graph, stmtNode, stmtXref, now); err != nil {
		return err
	}

	if err := processRepositoriesForSQLite(tree, graph, stmtNode, stmtXref, now); err != nil {
		return err
	}

	if err := processEventsForSQLite(tree, graph, stmtNode, stmtXref, now); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// processIndividualsForSQLite processes individual records for SQLite
func processIndividualsForSQLite(tree *gedcom.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, now int64) error {
	individuals := tree.GetAllIndividuals()

	for xrefID, record := range individuals {
		indiRecord, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

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

	return nil
}

// processFamiliesForSQLite processes family records for SQLite
func processFamiliesForSQLite(tree *gedcom.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, now int64) error {
	families := tree.GetAllFamilies()

	for xrefID, record := range families {
		_, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

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

	return nil
}

// processNotesForSQLite processes note records for SQLite
func processNotesForSQLite(tree *gedcom.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, now int64) error {
	notes := tree.GetAllNotes()

	for xrefID, record := range notes {
		_, ok := record.(*gedcom.NoteRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

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

	return nil
}

// processSourcesForSQLite processes source records for SQLite
func processSourcesForSQLite(tree *gedcom.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, now int64) error {
	sources := tree.GetAllSources()

	for xrefID, record := range sources {
		_, ok := record.(*gedcom.SourceRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

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

	return nil
}

// processRepositoriesForSQLite processes repository records for SQLite
func processRepositoriesForSQLite(tree *gedcom.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, now int64) error {
	repositories := tree.GetAllRepositories()

	for xrefID, record := range repositories {
		_, ok := record.(*gedcom.RepositoryRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

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

	return nil
}

// processEventsForSQLite processes event nodes for SQLite
func processEventsForSQLite(tree *gedcom.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, now int64) error {

	// Process individual events
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

			// Get or create node ID (with locking)
			graph.mu.Lock()
			nodeID := graph.xrefToID[eventID]
			if nodeID == 0 {
				nodeID = graph.nextID
				graph.nextID++
				graph.xrefToID[eventID] = nodeID
				graph.idToXref[nodeID] = eventID
			}
			graph.mu.Unlock()

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
	families := tree.GetAllFamilies()
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

