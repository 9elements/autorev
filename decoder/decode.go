package decode

import (
	"fmt"
	"github.com/9elements/autorev/tracelog"
)

func genCode(e TraceLogEntry) {
	entry := e.GetTraceLogEntry()

	Ip       string
	LineType int
	Inout    bool
	Address  string
	Value    string

	fmt.Printf("")
}

func MeshToHLSC(mesh Mesh) string {
	var list []*MeshNode = nil

	fmt.Printf("void main(void) {\n")

	list = mesh.Start.firstPath()

	for list != nil {
		for i, _ := range list {
			genCode(list[i])
		}
		list = mesh.Start.nextPath(list)
	}
	fmt.Printf("}\n")
}
