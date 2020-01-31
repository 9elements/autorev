package mesh

import (
	"testing"

	"github.com/9elements/autorev/tracelog"
)

func TestMeshAppendNode(t *testing.T) {
	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	n := MeshNode{Id: 1, Hash: "1"}
	m.appendNode(&n)

	if len(m.Nodes) != 1 {
		t.Errorf("Wrong node count in Nodes")
	}
	if len(m.Start.Next) != 1 {
		t.Errorf("Wrong node count on Start")
	}
	if len(m.Start.Next) >= 1 {
		if m.Start.Next[0] != &n {
			t.Errorf("Wrong node after start node")
		}
	}
	if len(n.Prev) != 1 {
		t.Errorf("Wrong node count on test node")
	}
	if len(n.Prev) >= 1 {
		if n.Prev[0] != &m.Start {
			t.Errorf("Wrong node on test node")
		}
	}
}

func TestFirstPath(t *testing.T) {
	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	a := MeshNode{Id: 1, Hash: "1"}
	b := MeshNode{Id: 2, Hash: "3"}
	c := MeshNode{Id: 3, Hash: "3"}
	d := MeshNode{Id: 4, Hash: "4"}

	m.appendNode(&a)
	m.appendNode(&b)
	m.appendNode(&c)
	m.insertNode(&a, &d)

	f1 := a.FirstPath()
	if len(f1) != 2 {
		t.Errorf("First path on a has wrong length %d", len(f1))
	}
	if len(f1) >= 2 {
		if f1[0] != &b {
			t.Errorf("First path on a has wrong first element")
		}
		if f1[1] != &c {
			t.Errorf("First path on a has wrong last element")
		}
	}
	f2 := b.FirstPath()
	if len(f2) != 1 {
		t.Errorf("First path on b has wrong length %d", len(f2))
	}
	if len(f1) >= 1 {
		if f2[0] != &c {
			t.Errorf("First path on b has wrong first element")
		}
	}
}

func TestNextPath(t *testing.T) {
	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	a := MeshNode{Id: 1, Hash: "1"}
	b := MeshNode{Id: 2, Hash: "2"}
	c := MeshNode{Id: 3, Hash: "3"}
	d := MeshNode{Id: 4, Hash: "4"}
	e := MeshNode{Id: 5, Hash: "5"}
	f := MeshNode{Id: 6, Hash: "6"}

	m.appendNode(&a)
	m.appendNode(&b)
	m.appendNode(&c)
	m.insertNode(&a, &d)
	m.insertNode(&d, &e)

	f1 := a.FirstPath()
	f2 := a.NextPath(f1)

	if len(f2) != 2 {
		t.Errorf("Next path on a has wrong length %d", len(f2))
	}
	if len(f2) >= 2 {
		if f2[0] != &d {
			t.Errorf("Next path on a has wrong first element")
		}
		if f2[1] != &e {
			t.Errorf("Next path on a has wrong last element")
		}
	}

	m.insertNode(&b, &f)
	f3 := a.FirstPath()
	f4 := a.NextPath(f3)

	if len(f4) != 2 {
		t.Errorf("Next path on a has wrong length %d", len(f4))
	}
	if len(f4) >= 2 {
		if f4[0] != &b {
			t.Errorf("Next path on a has wrong first element")
		}
		if f4[1] != &f {
			t.Errorf("Next path on a has wrong last element")
		}
	}
}

func TestMergeEmptyMesh(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "1"}}
	a := MeshNode{Id: 1, Hash: "1"}
	b.appendNode(&a)

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}

	merge(&b, &m)
	if len(m.Nodes) != 1 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Nodes) > 0 {
		if m.Nodes[0].Hash != a.Hash {
			t.Errorf("a hasn't been merged")
		}
	}
	if len(m.Start.Next) > 0 {
		if m.Start.Next[0].Hash != a.Hash {
			t.Errorf("a isn't next after Start")
		}
	}
}

func TestMergeEmptyBranch(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	a := MeshNode{Id: 1, Hash: "1"}
	m.appendNode(&a)

	merge(&b, &m)
	if len(m.Nodes) != 1 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Nodes) > 0 {
		if m.Nodes[0].Hash != a.Hash {
			t.Errorf("a hasn't been merged")
		}
	}
	if len(m.Start.Next) > 0 {
		if m.Start.Next[0].Hash != a.Hash {
			t.Errorf("a isn't next after Start")
		}
	}
}

func TestMergeEqualBranchMesh(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	var a = MeshNode{Id: 1, Hash: "1"}
	b.appendNode(&a)
	b.appendNode(&MeshNode{Id: 2, Hash: "2"})

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	m.appendNode(&MeshNode{Id: 1, Hash: "1"})
	m.appendNode(&MeshNode{Id: 2, Hash: "2"})

	merge(&b, &m)
	if len(m.Nodes) != 2 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Nodes) > 0 {
		if m.Nodes[0].Hash != a.Hash {
			t.Errorf("a hasn't been merged")
		}
	}
	if len(m.Start.Next) > 0 {
		if m.Start.Next[0].Hash != a.Hash {
			t.Errorf("a isn't next after Start")
		}
	}
	if m.Pathes != 1 {
		t.Errorf("m has wrong Pathes count %d", m.Pathes)
	}

	for i, _ := range m.Nodes {
		if m.Nodes[i].Propability != 1 {
			t.Errorf("Node %d on m has wrong Propability count %d", i, m.Nodes[i].Propability)
		}
	}
}

func TestMergeSimple(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	b.appendNode(&MeshNode{Id: 1, Hash: "1"})
	b.appendNode(&MeshNode{Id: 2, Hash: "2"})
	b.appendNode(&MeshNode{Id: 3, Hash: "3"})

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	var a = MeshNode{Id: 1, Hash: "1"}
	m.appendNode(&a)
	m.appendNode(&MeshNode{Id: 4, Hash: "4"})
	var c = MeshNode{Id: 3, Hash: "3"}
	m.appendNode(&c)

	if len(c.Prev) != 1 {
		t.Errorf("Wrong connection count on c node in mesh: %d", len(c.Next))
	}

	merge(&b, &m)

	if len(m.Nodes) != 4 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Nodes) > 0 {
		if m.Nodes[0].Hash != a.Hash {
			t.Errorf("a hasn't been merged")
		}
	}
	if len(m.Start.Next) != 1 {
		t.Errorf("Wrong connection count on Start node")
	}
	if len(a.Next) != 2 {
		t.Errorf("Wrong connection count on a node in mesh: %d", len(a.Next))
	}
	if len(c.Prev) != 2 {
		t.Errorf("Wrong connection count on c node in mesh: %d", len(c.Prev))
	}
}

func TestMergeLoseEnd(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	b.appendNode(&MeshNode{Id: 1, Hash: "1"})
	b.appendNode(&MeshNode{Id: 2, Hash: "2"})
	b.appendNode(&MeshNode{Id: 3, Hash: "3"})
	b.appendNode(&MeshNode{Id: 4, Hash: "4"})
	b.appendNode(&MeshNode{Id: 5, Hash: "5"})
	b.appendNode(&MeshNode{Id: 6, Hash: "6"})

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	var a = MeshNode{Id: 1, Hash: "1"}
	m.appendNode(&a)
	m.appendNode(&MeshNode{Id: 4, Hash: "4"})
	var c = MeshNode{Id: 3, Hash: "3"}
	m.appendNode(&c)
	m.appendNode(&MeshNode{Id: 7, Hash: "7"})
	m.appendNode(&MeshNode{Id: 8, Hash: "8"})
	m.appendNode(&MeshNode{Id: 9, Hash: "9"})

	if len(c.Prev) != 1 {
		t.Errorf("Wrong connection count on c node in mesh: %d", len(c.Next))
	}

	merge(&b, &m)

	if len(m.Nodes) != 10 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Nodes) > 0 {
		if m.Nodes[0].Hash != a.Hash {
			t.Errorf("a hasn't been merged")
		}
	}
	if len(m.Start.Next) != 1 {
		t.Errorf("Wrong connection count on Start node")
	}
	if len(a.Next) != 2 {
		t.Errorf("Wrong next connection count on a node in mesh: %d", len(a.Next))
	}
	if len(c.Prev) != 2 {
		t.Errorf("Wrong prev connection count on c node in mesh: %d", len(c.Prev))
	}
	if len(c.Next) != 2 {
		t.Errorf("Wrong next connection count on c node in mesh: %d", len(c.Next))
	}
}

func TestMergeComplexA(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	b.appendNode(&MeshNode{Id: 1, Hash: "1"})
	b.appendNode(&MeshNode{Id: 2, Hash: "2"})
	b.appendNode(&MeshNode{Id: 3, Hash: "3"})
	b.appendNode(&MeshNode{Id: 4, Hash: "4"})
	b.appendNode(&MeshNode{Id: 5, Hash: "5"})
	b.appendNode(&MeshNode{Id: 6, Hash: "6"})
	b.appendNode(&MeshNode{Id: 7, Hash: "7"})
	b.appendNode(&MeshNode{Id: 8, Hash: "8"})
	b.appendNode(&MeshNode{Id: 9, Hash: "9"})

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	var a = MeshNode{Id: 1, Hash: "1"}
	m.appendNode(&a)
	var c = MeshNode{Id: 4, Hash: "4"}
	m.appendNode(&c)
	m.appendNode(&MeshNode{Id: 5, Hash: "5"})
	m.appendNode(&MeshNode{Id: 6, Hash: "6"})
	var d = MeshNode{Id: 7, Hash: "7"}
	m.appendNode(&d)
	m.appendNode(&MeshNode{Id: 11, Hash: "11"})
	m.appendNode(&MeshNode{Id: 9, Hash: "9"})

	if len(c.Prev) != 1 {
		t.Errorf("Wrong connection count on c node in mesh: %d", len(c.Next))
	}

	merge(&b, &m)

	if len(m.Nodes) != 10 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Nodes) > 0 {
		if m.Nodes[0].Hash != a.Hash {
			t.Errorf("a hasn't been merged")
		}
	}
	if len(m.Start.Next) != 1 {
		t.Errorf("Wrong connection count on Start node")
	}
	if len(a.Next) != 2 {
		t.Errorf("Wrong next connection count on a node in mesh: %d", len(a.Next))
	}
	if len(c.Prev) != 2 {
		t.Errorf("Wrong prev connection count on c node in mesh: %d", len(c.Prev))
	}
	if len(d.Next) != 2 {
		t.Errorf("Wrong next connection count on d node in mesh: %d", len(d.Next))
	}
}

func TestMergeComplexB(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	b.appendNode(&MeshNode{Id: 1, Hash: "1"})
	b.appendNode(&MeshNode{Id: 4, Hash: "4"})
	b.appendNode(&MeshNode{Id: 5, Hash: "5"})
	b.appendNode(&MeshNode{Id: 6, Hash: "6"})
	b.appendNode(&MeshNode{Id: 7, Hash: "7"})
	b.appendNode(&MeshNode{Id: 8, Hash: "8"})
	b.appendNode(&MeshNode{Id: 9, Hash: "9"})

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	var a = MeshNode{Id: 1, Hash: "1"}
	m.appendNode(&a)
	m.appendNode(&MeshNode{Id: 2, Hash: "2"})
	m.appendNode(&MeshNode{Id: 3, Hash: "3"})
	var c = MeshNode{Id: 4, Hash: "4"}
	m.appendNode(&c)
	m.appendNode(&MeshNode{Id: 5, Hash: "5"})
	m.appendNode(&MeshNode{Id: 6, Hash: "6"})
	var d = MeshNode{Id: 7, Hash: "7"}
	m.appendNode(&d)
	m.appendNode(&MeshNode{Id: 11, Hash: "11"})
	m.appendNode(&MeshNode{Id: 9, Hash: "9"})

	if len(c.Prev) != 1 {
		t.Errorf("Wrong connection count on c node in mesh: %d", len(c.Next))
	}

	merge(&b, &m)

	if len(m.Nodes) != 10 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Nodes) > 0 {
		if m.Nodes[0].Hash != a.Hash {
			t.Errorf("a hasn't been merged")
		}
	}
	if len(m.Start.Next) != 1 {
		t.Errorf("Wrong connection count on Start node")
	}
	if len(a.Next) != 2 {
		t.Errorf("Wrong next connection count on a node in mesh: %d", len(a.Next))
	}
	if len(c.Prev) != 2 {
		t.Errorf("Wrong prev connection count on c node in mesh: %d", len(c.Prev))
	}
	if len(d.Next) != 2 {
		t.Errorf("Wrong next connection count on d node in mesh: %d", len(d.Next))
	}
}

func TestInsertTraceLogIntoMeshA(t *testing.T) {
	var tle = []tracelog.TraceLogEntry{
		tracelog.TraceLogEntry{
			IP:         1,
			Type:       0,
			Inout:      true,
			Address:    0xdeafbeef,
			Value:      0,
			AccessSize: 8},
		tracelog.TraceLogEntry{
			IP:         1,
			Type:       0,
			Inout:      true,
			Address:    0xdeafbeef,
			Value:      0,
			AccessSize: 8},
		tracelog.TraceLogEntry{
			IP:         1,
			Type:       0,
			Inout:      true,
			Address:    0xdeafbeef,
			Value:      0,
			AccessSize: 16},
		tracelog.TraceLogEntry{
			IP:         2,
			Type:       0,
			Inout:      true,
			Address:    0xdeafbeef,
			Value:      0,
			AccessSize: 8},
	}
	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}, ID: 1}

	m.InsertTraceLogIntoMesh(tle, nil)

	if len(m.Nodes) != 4 {
		t.Errorf("Wrong node count in mesh: %d", len(m.Nodes))
		return
	}

	if m.Start.Next[0].Hash != m.Start.Next[0].Next[0].Hash {
		t.Errorf("Hash of first and second Nodes doesn't match")
	}
	if m.Start.Next[0].Hash == m.Start.Next[0].Next[0].Next[0].Hash {
		t.Errorf("Hash of first and third Nodes do match")
	}
	if m.Start.Next[0].Hash == m.Start.Next[0].Next[0].Next[0].Next[0].Hash {
		t.Errorf("Hash of first and forth Nodes do match")
	}
}

func TestInsertTraceLogIntoMeshB(t *testing.T) {
	var m = Mesh{Start: MeshNode{Id: 0}, ID: 1}

	var tles1 = []tracelog.TraceLogEntry{
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef1, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef2, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef3, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef, AccessSize: 16},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef, AccessSize: 32},
	}

	var tles2 = []tracelog.TraceLogEntry{
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef1, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef2, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef3, AccessSize: 8},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef, AccessSize: 32},
		{IP: 1, Type: 0, Inout: false, Address: 0, Value: 0xdeadbeef, AccessSize: 32},
	}

	err := m.InsertTraceLogIntoMesh(tles1, nil)
	if err != nil {
		t.Errorf("Error generating branch from tracelog entries: %v", tles1)
	}

	err = m.InsertTraceLogIntoMesh(tles2, nil)
	if err != nil {
		t.Errorf("Error generating branch from tracelog entries: %v", tles2)
	}

	if len(m.Nodes) != 7 {
		t.Errorf("Unexpcted node count: %d", len(m.Nodes))
	}

	if len(m.Start.Next[0].Next[0].Next[0].Next[0].Next) != 2 {
		t.Errorf("Unexpected next connection count on 4th node: %d", len(m.Start.Next[0].Next[0].Next[0].Next[0].Next))
	}
	if len(m.Start.Next[0].Next[0].Next[0].Next[0].Next[0].Next[0].Prev) != 2 {
		t.Errorf("Unexpected prev connection count on 6th node: %d", len(m.Start.Next[0].Next[0].Next[0].Next[0].Next[0].Next[0].Prev))
	}
}

func TestFirmwareOptionMerge(t *testing.T) {
	var b = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	b.appendNode(&MeshNode{Id: 1, Hash: "1", FirmwareOptions: []map[string]uint64{map[string]uint64{"OPTION1": 1, "OPTION2": 2}, map[string]uint64{"OPTION3": 1}}})

	var m = Mesh{Start: MeshNode{Id: 0, Hash: "0"}}
	var a = MeshNode{Id: 1, Hash: "1", FirmwareOptions: []map[string]uint64{map[string]uint64{"OPTION1": 1, "OPTION2": 2}, map[string]uint64{"OPTION4": 1}}}
	m.appendNode(&a)

	merge(&b, &m)

	if len(m.Nodes) != 1 {
		t.Errorf("Wrong node count in final mesh: %d", len(m.Nodes))
	}
	if len(m.Start.Next[0].FirmwareOptions) != 3 {
		t.Errorf("Wrong entry count on first node for FirmwareOption OPTION1: %d", len(m.Start.Next[0].FirmwareOptions))
	}
	for i := range m.Start.Next[0].FirmwareOptions {
		if m.Start.Next[0].FirmwareOptions[i]["OPTION1"] > 0 {
			if m.Start.Next[0].FirmwareOptions[i]["OPTION1"] != 1 {
				t.Errorf("Wrong FirmwareOption OPTION1: %d", m.Start.Next[0].FirmwareOptions[i]["OPTION1"])
			}
			if m.Start.Next[0].FirmwareOptions[i]["OPTION2"] != 2 {
				t.Errorf("Wrong FirmwareOption OPTION2: %d", m.Start.Next[0].FirmwareOptions[i]["OPTION2"])
			}
			if m.Start.Next[0].FirmwareOptions[i]["OPTION3"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION3 in map")
			}
			if m.Start.Next[0].FirmwareOptions[i]["OPTION4"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION4 in map")
			}
		}
		if m.Start.Next[0].FirmwareOptions[i]["OPTION3"] > 0 {
			if m.Start.Next[0].FirmwareOptions[i]["OPTION1"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION1 in map")
			}
			if m.Start.Next[0].FirmwareOptions[i]["OPTION2"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION2 in map")
			}
			if m.Start.Next[0].FirmwareOptions[i]["OPTION4"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION4 in map")
			}
		}
		if m.Start.Next[0].FirmwareOptions[i]["OPTION4"] > 0 {
			if m.Start.Next[0].FirmwareOptions[i]["OPTION1"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION1 in map")
			}
			if m.Start.Next[0].FirmwareOptions[i]["OPTION2"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION2 in map")
			}
			if m.Start.Next[0].FirmwareOptions[i]["OPTION3"] > 0 {
				t.Errorf("Wrong FirmwareOption OPTION3 in map")
			}
		}
	}
}
