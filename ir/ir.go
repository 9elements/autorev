package ir

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/9elements/autorev/mesh"
	"github.com/9elements/autorev/tracelog"
)

type IRType int

const (
	PrimitiveRead IRType = iota
	PrimitiveWrite
	ComplexReadModifyWrite
)

type PrimtiveType int

const (
	MEM32 PrimtiveType = iota
	IO
	MSR
	CPUID
	PCI
)

type ir interface {
	GetType() IRType
	GetRank() uint
	ConvertToC() string
}

type IRMeta struct {
	hash uint
}

type PRead struct {
	// Meta information
	Meta IRMeta
	// The primitive type (TraceLogEntry Type)
	Type PrimtiveType
	// The address that was accessed
	Address uint
	// The value read/written
	Value uint64
	// 8, 16, 32, 64 bit
	AccessSize uint
}

func (p PRead) GetType() IRType {
	return PrimitiveRead
}

func (p PRead) GetRank() uint {
	return 1
}

type PWrite struct {
	// Meta information
	Meta IRMeta
	// The primitive type (TraceLogEntry Type)
	Type PrimtiveType
	// The address that was accessed
	Address uint
	// The value read/written
	Value uint64
	// 8, 16, 32, 64 bit
	AccessSize uint
}

func (p PWrite) GetType() IRType {
	return PrimitiveWrite
}

func (p PWrite) GetRank() uint {
	return 1
}

type CRMW struct {
	Type PrimtiveType
	// The address that was accessed
	Address uint
	// The OrMask to set bits
	OrMask uint64
	// The AndMask to clear bits
	AndMask uint64
	// 8, 16, 32, 64 bit
	AccessSize uint
}

func (c CRMW) GetType() IRType {
	return ComplexReadModifyWrite
}

func (c CRMW) GetRank() uint {
	return 2
}

func IRNewPrimitiveRead(e tracelog.TraceLogEntry) PRead {
	var p PRead
	p.AccessSize = e.AccessSize
	p.Address = e.Address
	p.Value = e.Value
	if e.Type == int(tracelog.IO) {
		p.Type = IO
	} else if e.Type == int(tracelog.CPUID) {
		p.Type = CPUID
	} else if e.Type == int(tracelog.MSR) {
		p.Type = MSR
	} else if e.Type == int(tracelog.PCI) {
		p.Type = PCI
	} else if e.Type == int(tracelog.MEM32) {
		p.Type = MEM32
	}
	return p
}

func IRNewPrimitiveWrite(e tracelog.TraceLogEntry) PWrite {
	var p PWrite
	p.AccessSize = e.AccessSize
	p.Address = e.Address
	p.Value = e.Value
	if e.Type == int(tracelog.IO) {
		p.Type = IO
	} else if e.Type == int(tracelog.CPUID) {
		p.Type = CPUID
	} else if e.Type == int(tracelog.MSR) {
		p.Type = MSR
	} else if e.Type == int(tracelog.PCI) {
		p.Type = PCI
	} else if e.Type == int(tracelog.MEM32) {
		p.Type = MEM32
	}
	return p
}

func PrimitiveToIR(n *mesh.MeshNode) string {
	var ret string
	if n.IsNoop {
		return ""
	}
	if n.TLE.Inout {
		p := IRNewPrimitiveRead(n.TLE)
		line := fmt.Sprintf("%s", p.ConvertToC())
		ret += line
	} else {
		p := IRNewPrimitiveWrite(n.TLE)
		line := fmt.Sprintf("%s", p.ConvertToC())
		ret += line
	}
	return ret
}

func LoopToIR(ident int, start *mesh.MeshNode, end *mesh.MeshNode) string {
	ret := ""
	whitespace := strings.Repeat(" ", ident*2)
	n := start

	for true {
		ret += whitespace + PrimitiveToIR(n)
		if len(n.Next) == 0 {
			return ret
		} else if len(n.Next) == 1 {
			n = n.Next[0]
		} else {
			// Get the merge point that contains all "loops"
			mergeNode := n.CommonMergePoint()
			if mergeNode == nil {
				fmt.Errorf("No MergePoint found. FIXME: Implement support for dead-ends!")
				return ret
			}
			// Get a condition from one of the branches
			branchWithCond := -1
			for i := range n.Next {
				// FIXME: sort nocondition
				cond := false
				for u := range n.Next[i].FirmwareOptions {
					if len(n.Next[i].FirmwareOptions[u]) > 0 {
						cond = true
						break
					}
				}
				if cond {
					branchWithCond = i
					break
				}
			}
			if branchWithCond == -1 {
				fmt.Errorf("Branch detected but no condition found!")
				return ret
			}

			// Emit a branch with a condition
			for i := branchWithCond; i == branchWithCond; i++ {
				ret += whitespace + "if ("
				isfirstcond := true

				for u := range n.Next[i].FirmwareOptions {
					if !isfirstcond {
						ret += " ||\n" + whitespace + "    "
					}
					isfirstcond = false
					ret += "("
					isfirst := true
					for k, v := range n.Next[i].FirmwareOptions[u] {
						if !isfirst {
							ret += " && "
						}
						ret += k + " == " + strconv.FormatUint(v, 10)
						isfirst = false

					}
					ret += ")"
				}
				ret += ") {\n"

				ret += LoopToIR(ident+1, n.Next[i], mergeNode)
				ret += whitespace + "}\n"
			}

			// Now emit all else if, else branches
			for i := range n.Next {
				if i == branchWithCond {
					continue
				}
				nocond := true
				for u := range n.Next[i].FirmwareOptions {
					if len(n.Next[i].FirmwareOptions[u]) > 0 {
						nocond = false
						break
					}
				}
				if !nocond {
					ret += whitespace + "else if ("
					isfirstcond := true

					for u := range n.Next[i].FirmwareOptions {
						if !isfirstcond {
							ret += " ||\n" + whitespace + "    "
						}
						isfirstcond = false
						ret += "("
						isfirst := true
						for k, v := range n.Next[i].FirmwareOptions[u] {
							if !isfirst {
								ret += " && "
							}
							ret += k + " == " + strconv.FormatUint(v, 10)
							isfirst = false

						}
						ret += ")"
					}
					ret += ") {\n"
				} else {
					ret += whitespace + "else {\n"
				}

				ret += LoopToIR(ident+1, n.Next[i], mergeNode)
				ret += whitespace + "}\n"
			}

			// find branch merge point
			if mergeNode != nil {
				n = mergeNode
			} else {
				fmt.Errorf("No MergePoint found. FIXME: Implement support for deadends!")
				return ret
			}
		}
		if n.Id == end.Id {
			return ret
		}
	}
	return ret
}

func MeshToIR(m *mesh.Mesh) string {
	return LoopToIR(0, &m.Start, m.LastNode())
}
