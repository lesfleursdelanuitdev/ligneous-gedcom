package query

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// NodeData represents serialized node data for BadgerDB
// Note: We don't serialize the full record - we store the XREF and reconstruct from tree
// For EventNode, we also store EventType and EventData since events don't have top-level records
type NodeData struct {
	ID       uint32
	Xref     string
	NodeType NodeType
	// Event-specific fields (only used for NodeTypeEvent)
	EventType string
	EventData map[string]interface{}
	// Data field removed - we'll reconstruct from tree using Xref
}

// EdgeData represents serialized edge data for BadgerDB
type EdgeData struct {
	FromID     uint32
	ToID       uint32
	EdgeType   EdgeType
	FamilyID   uint32 // For FAMC/FAMS edges
	Direction  Direction
	Properties map[string]interface{}
}

// SerializeNode serializes a node to bytes for storage in BadgerDB
// Note: We only store metadata (XREF, type) - the full record is reconstructed from tree
// For EventNode, we also store EventType and EventData
func SerializeNode(node GraphNode, graph *Graph) ([]byte, error) {
	nodeData := NodeData{
		ID:       getNodeIDFromXref(node.ID(), graph),
		Xref:     node.ID(),
		NodeType: node.NodeType(),
	}

	// For EventNode, store event-specific data
	if eventNode, ok := node.(*EventNode); ok {
		nodeData.EventType = eventNode.EventType
		nodeData.EventData = eventNode.EventData
	}

	return serialize(nodeData)
}

// DeserializeNode deserializes bytes from BadgerDB into a node
// Reconstructs the record from the GEDCOM tree using the XREF
func DeserializeNode(data []byte, graph *Graph) (GraphNode, error) {
	var nodeData NodeData
	if err := deserialize(data, &nodeData); err != nil {
		return nil, fmt.Errorf("failed to deserialize node data: %w", err)
	}

	// Reconstruct record from tree using XREF
	var record types.Record
	switch nodeData.NodeType {
	case NodeTypeIndividual:
		record = graph.tree.GetIndividual(nodeData.Xref)
		if record == nil {
			return nil, fmt.Errorf("individual %s not found in tree", nodeData.Xref)
		}
		indiRecord, ok := record.(*types.IndividualRecord)
		if !ok {
			return nil, fmt.Errorf("expected IndividualRecord, got %T", record)
		}
		return &IndividualNode{
			BaseNode: &BaseNode{
				xrefID:   nodeData.Xref,
				nodeType: NodeTypeIndividual,
				record:   indiRecord,
			},
			Individual: indiRecord,
		}, nil

	case NodeTypeFamily:
		record = graph.tree.GetFamily(nodeData.Xref)
		if record == nil {
			return nil, fmt.Errorf("family %s not found in tree", nodeData.Xref)
		}
		famRecord, ok := record.(*types.FamilyRecord)
		if !ok {
			return nil, fmt.Errorf("expected FamilyRecord, got %T", record)
		}
		return &FamilyNode{
			BaseNode: &BaseNode{
				xrefID:   nodeData.Xref,
				nodeType: NodeTypeFamily,
				record:   famRecord,
			},
			Family: famRecord,
		}, nil

	case NodeTypeNote:
		// Get note from tree's xref index
		record = graph.tree.GetRecordByXref(nodeData.Xref)
		if record == nil {
			return nil, fmt.Errorf("note %s not found in tree", nodeData.Xref)
		}
		noteRecord, ok := record.(*types.NoteRecord)
		if !ok {
			return nil, fmt.Errorf("expected NoteRecord, got %T", record)
		}
		return &NoteNode{
			BaseNode: &BaseNode{
				xrefID:   nodeData.Xref,
				nodeType: NodeTypeNote,
				record:   noteRecord,
			},
			Note: noteRecord,
		}, nil

	case NodeTypeSource:
		// Get source from tree's xref index
		record = graph.tree.GetRecordByXref(nodeData.Xref)
		if record == nil {
			return nil, fmt.Errorf("source %s not found in tree", nodeData.Xref)
		}
		sourceRecord, ok := record.(*types.SourceRecord)
		if !ok {
			return nil, fmt.Errorf("expected SourceRecord, got %T", record)
		}
		return &SourceNode{
			BaseNode: &BaseNode{
				xrefID:   nodeData.Xref,
				nodeType: NodeTypeSource,
				record:   sourceRecord,
			},
			Source: sourceRecord,
		}, nil

	case NodeTypeRepository:
		// Get repository from tree's xref index
		record = graph.tree.GetRecordByXref(nodeData.Xref)
		if record == nil {
			return nil, fmt.Errorf("repository %s not found in tree", nodeData.Xref)
		}
		repoRecord, ok := record.(*types.RepositoryRecord)
		if !ok {
			return nil, fmt.Errorf("expected RepositoryRecord, got %T", record)
		}
		return &RepositoryNode{
			BaseNode: &BaseNode{
				xrefID:   nodeData.Xref,
				nodeType: NodeTypeRepository,
				record:   repoRecord,
			},
			Repository: repoRecord,
		}, nil

	case NodeTypeEvent:
		// Events don't have top-level records - reconstruct from stored EventType and EventData
		if nodeData.EventType == "" {
			return nil, fmt.Errorf("event %s missing EventType", nodeData.Xref)
		}
		if nodeData.EventData == nil {
			nodeData.EventData = make(map[string]interface{})
		}
		// Ensure event type is in event data
		if nodeData.EventData["type"] == nil {
			nodeData.EventData["type"] = nodeData.EventType
		}
		return &EventNode{
			BaseNode: &BaseNode{
				xrefID:   nodeData.Xref,
				nodeType: NodeTypeEvent,
				record:   nil, // Events don't have top-level records
				inEdges:  make([]*Edge, 0),
				outEdges: make([]*Edge, 0),
			},
			EventID:   nodeData.Xref,
			EventType: nodeData.EventType,
			EventData: nodeData.EventData,
			Sources:   make([]*SourceNode, 0),
			Notes:     make([]*NoteNode, 0),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeData.NodeType)
	}
}

// SerializeEdges serializes a slice of edges to bytes
func SerializeEdges(edges []*Edge, graph *Graph) ([]byte, error) {
	edgeDataList := make([]EdgeData, 0, len(edges))
	for _, edge := range edges {
		edgeData := EdgeData{
			FromID:     getNodeIDFromXref(edge.From.ID(), graph),
			ToID:       getNodeIDFromXref(edge.To.ID(), graph),
			EdgeType:   edge.EdgeType,
			Direction:  edge.Direction,
			Properties: edge.Properties,
		}

		// Extract family ID if this is a family-related edge
		if edge.Family != nil {
			edgeData.FamilyID = getNodeIDFromXref(edge.Family.ID(), graph)
		}

		edgeDataList = append(edgeDataList, edgeData)
	}

	return serialize(edgeDataList)
}

// serializeEdgeDataList serializes a slice of EdgeData directly (for use during construction)
func serializeEdgeDataList(edges []EdgeData) ([]byte, error) {
	return serialize(edges)
}

// DeserializeEdges deserializes bytes into a slice of edges
func DeserializeEdges(data []byte, graph *Graph) ([]*Edge, error) {
	var edgeDataList []EdgeData
	if err := deserialize(data, &edgeDataList); err != nil {
		return nil, fmt.Errorf("failed to deserialize edges: %w", err)
	}

	edges := make([]*Edge, 0, len(edgeDataList))
	for _, edgeData := range edgeDataList {
		fromXref := graph.idToXref[edgeData.FromID]
		toXref := graph.idToXref[edgeData.ToID]

		fromNode := graph.GetNode(fromXref)
		toNode := graph.GetNode(toXref)

		if fromNode == nil || toNode == nil {
			continue // Skip if nodes not found
		}

		edge := &Edge{
			ID:         fmt.Sprintf("%s_%s_%s", fromXref, string(edgeData.EdgeType), toXref),
			From:       fromNode,
			To:         toNode,
			EdgeType:   edgeData.EdgeType,
			Direction:  edgeData.Direction,
			Properties: edgeData.Properties,
		}

		// Load family node if needed
		if edgeData.FamilyID != 0 {
			famXref := graph.idToXref[edgeData.FamilyID]
			if famNode := graph.GetFamily(famXref); famNode != nil {
				edge.Family = famNode
			}
		}

		edges = append(edges, edge)
	}

	return edges, nil
}

// SerializeComponents serializes component data (node IDs) to bytes
func SerializeComponents(nodeIDs []uint32) ([]byte, error) {
	return serialize(nodeIDs)
}

// DeserializeComponents deserializes component data from bytes
func DeserializeComponents(data []byte) ([]uint32, error) {
	var nodeIDs []uint32
	if err := deserialize(data, &nodeIDs); err != nil {
		return nil, fmt.Errorf("failed to deserialize components: %w", err)
	}
	return nodeIDs, nil
}

// Helper functions for serialization

func serialize(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("gob encode error: %w", err)
	}
	return buf.Bytes(), nil
}

func deserialize(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("gob decode error: %w", err)
	}
	return nil
}

// Note: serializeRecord and deserializeRecord are no longer needed
// We store only XREF and reconstruct from tree

// Helper to get node ID from XREF using graph's mapping
func getNodeIDFromXref(xref string, graph *Graph) uint32 {
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	return graph.xrefToID[xref]
}

