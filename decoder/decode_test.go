package decode

import (
	"testing"

	"github.com/9elements/autorev/mesh"
)

func TestMeshToHLSC(t *testing.T) {
	var b = mesh.Mesh{Start: mesh.MeshNode{Id: 0}}
	b.AppendNode(&mesh.MeshNode{Id: 1})
	b.AppendNode(&mesh.MeshNode{Id: 2})
	b.AppendNode(&mesh.MeshNode{Id: 3})
	b.AppendNode(&mesh.MeshNode{Id: 4})
	b.AppendNode(&mesh.MeshNode{Id: 5})

	s := MeshToHLSC(b)
	t.Logf(s)
}
