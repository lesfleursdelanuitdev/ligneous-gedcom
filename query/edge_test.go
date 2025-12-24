package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestNewEdge(t *testing.T) {
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	node1 := NewIndividualNode("@I1@", indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := types.NewIndividualRecord(indi2Line)
	node2 := NewIndividualNode("@I2@", indi2)

	edge := NewEdge("E1", node1, node2, EdgeTypeSpouse)

	if edge.ID != "E1" {
		t.Errorf("Expected edge ID E1, got %s", edge.ID)
	}

	if edge.From != node1 {
		t.Error("Expected From to be node1")
	}

	if edge.To != node2 {
		t.Error("Expected To to be node2")
	}

	if edge.EdgeType != EdgeTypeSpouse {
		t.Errorf("Expected EdgeTypeSpouse, got %s", edge.EdgeType)
	}

	if edge.Direction != DirectionForward {
		t.Errorf("Expected DirectionForward, got %s", edge.Direction)
	}

	if edge.Weight != 1.0 {
		t.Errorf("Expected weight 1.0, got %f", edge.Weight)
	}
}

func TestNewBidirectionalEdge(t *testing.T) {
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	node1 := NewIndividualNode("@I1@", indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := types.NewIndividualRecord(indi2Line)
	node2 := NewIndividualNode("@I2@", indi2)

	edge := NewBidirectionalEdge("E1", node1, node2, EdgeTypeSpouse)

	if !edge.IsBidirectional() {
		t.Error("Expected edge to be bidirectional")
	}

	if edge.Direction != DirectionBidirectional {
		t.Errorf("Expected DirectionBidirectional, got %s", edge.Direction)
	}
}

func TestEdge_Connects(t *testing.T) {
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	node1 := NewIndividualNode("@I1@", indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := types.NewIndividualRecord(indi2Line)
	node2 := NewIndividualNode("@I2@", indi2)

	edge := NewEdge("E1", node1, node2, EdgeTypeSpouse)

	if !edge.Connects(node1) {
		t.Error("Expected edge to connect node1")
	}

	if !edge.Connects(node2) {
		t.Error("Expected edge to connect node2")
	}

	// Test with different node
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3 := types.NewIndividualRecord(indi3Line)
	node3 := NewIndividualNode("@I3@", indi3)

	if edge.Connects(node3) {
		t.Error("Expected edge not to connect node3")
	}
}

func TestEdge_OtherNode(t *testing.T) {
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	node1 := NewIndividualNode("@I1@", indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := types.NewIndividualRecord(indi2Line)
	node2 := NewIndividualNode("@I2@", indi2)

	edge := NewEdge("E1", node1, node2, EdgeTypeSpouse)

	other := edge.OtherNode(node1)
	if other != node2 {
		t.Error("Expected OtherNode(node1) to return node2")
	}

	other = edge.OtherNode(node2)
	if other != node1 {
		t.Error("Expected OtherNode(node2) to return node1")
	}
}

func TestNewEdgeWithFamily(t *testing.T) {
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := types.NewIndividualRecord(indiLine)
	node := NewIndividualNode("@I1@", indi)

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam := types.NewFamilyRecord(famLine)
	famNode := NewFamilyNode("@F1@", fam)

	edge := NewEdgeWithFamily("E1", node, famNode, EdgeTypeFAMC, famNode)

	if edge.Family != famNode {
		t.Error("Expected Family to be set")
	}
}
