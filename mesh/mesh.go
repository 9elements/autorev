package mesh

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/9elements/autorev/tracelog"

	"github.com/emicklei/dot"
	lcs "github.com/yudai/golcs"
)

// MeshNode - Describes a node in the Mesh
type MeshNode struct {
	// Id is incremented for each node added. It's unique in the mesh.
	Id uint64
	// Propability is incremented with each merged path if the node is part
	// of the merged path. By dividing it with Mesh.Pathes it gives the
	// propability
	Propability     uint64
	Next, Prev      []*MeshNode
	Hash            string
	TLE             tracelog.TraceLogEntry
	FirmwareOptions []map[string]uint64
	// a Noop meshnode doesn't generate code. It just makes the code generation prettier
	IsNoop bool
}

// a Mesh connects Nodes, using one or more pathes
type Mesh struct {
	Start MeshNode
	// Pathes is incremented on each merge
	Pathes uint64
	Nodes  []*MeshNode
	// IDcounter, increment on new MeshNode
	ID uint64
}

// a Branch is a Mesh, but only has one path
type Branch Mesh

func lcsMeshNodes(left []*MeshNode, right []*MeshNode) []interface{} {
	var leftIface []interface{} = make([]interface{}, len(left))
	var rightIface []interface{} = make([]interface{}, len(right))

	for i, d := range left {
		leftIface[i] = d.Hash
	}
	for i, d := range right {
		rightIface[i] = d.Hash
	}

	return lcs.New(leftIface, rightIface).Values()
}

// comparePrev- Compares Previous nodes if they are equal
func (mn *MeshNode) comparePrev(other *MeshNode) bool {
	if len(mn.Prev) != len(other.Prev) {
		return false
	}

	for _, v1 := range mn.Prev {
		found := false
		for _, v2 := range other.Prev {
			if v1.Id == v2.Id {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}

// returns likliest path starting at the children of the given node
func (mn *MeshNode) likelyPath() []*MeshNode {
	i := mn
	var list []*MeshNode

	for true {
		if len(i.Next) == 0 {
			break
		}
		var p uint64
		var l *MeshNode
		for j := range i.Next {
			if i.Next[j].Propability >= p {
				p = i.Next[j].Propability
				l = i.Next[j]
			}
		}
		// use the one with the highest propability
		i = l
		list = append(list, i)
	}
	return list
}

// FirstPath - returns first path starting at the children of the given node
func (mn *MeshNode) FirstPath() []*MeshNode {
	i := mn
	var list []*MeshNode

	for true {
		if len(i.Next) == 0 {
			break
		}
		i = i.Next[0]
		list = append(list, i)
	}
	return list
}

// LenFirstPath - Returns the length of the first path through the mesh
func (mn *MeshNode) LenFirstPath() uint {
	i := mn
	length := uint(0)

	for true {
		if len(i.Next) == 0 {
			break
		}
		i = i.Next[0]
		length++
	}
	return length
}

// NextPath - returns the next path starting at the children of the given node.
// You have to pass the last path (returned by FirstPath or NextPath) as argument
func (mn *MeshNode) NextPath(lastPath []*MeshNode) []*MeshNode {
	var last *MeshNode
	var list []*MeshNode

	if len(lastPath) <= 1 {
		return nil
	}

	lastPath = append([]*MeshNode{mn}, lastPath...)

	// set last element
	last = lastPath[len(lastPath)-1]

	// remove one from queue
	lastPath = lastPath[:len(lastPath)-1]

	for len(lastPath) > 0 {
		n := lastPath[len(lastPath)-1]
		if len(n.Next) > 1 {
			// branch
			found := -1
			for idx := range n.Next {
				if n.Next[idx].Hash == last.Hash {
					found = idx
					break
				}
			}
			if found == -1 {
				// error
				return nil
			}
			// more branches?
			if found+1 != len(n.Next) {
				list = append(lastPath, n.Next[found+1])
				break
			}
		}
		last = lastPath[len(lastPath)-1]
		lastPath = lastPath[:len(lastPath)-1]
	}

	if list == nil {
		return list
	}

	// cut of first element
	list = list[1:]

	n := list[len(list)-1]
	for len(n.Next) > 0 {
		list = append(list, n.Next[0])
		n = n.Next[0]
	}
	return list
}

// AnyPathesContainNode - returns true if one of the pathes contains the meshnode to find
func (mn *MeshNode) AnyPathesContainNode(find *MeshNode) bool {
	for p := mn.FirstPath(); p != nil; p = mn.NextPath(p) {
		for _, i := range p {
			if i.Id == find.Id {
				return true
			}
		}
	}
	return false
}

// CommonMergePoint - returns the merge point of all branches starting from this node
func (mn *MeshNode) CommonMergePoint() *MeshNode {
	var n *MeshNode
	for i, first := range mn.Next {
		for j := i; j < len(mn.Next); j++ {
			second := mn.Next[j]
			for left := first.FirstPath(); left != nil; left = first.NextPath(left) {
				for right := second.FirstPath(); right != nil; right = first.NextPath(right) {
					/* Compare all pathes between left and right and find the mergepoint */
					for i := 0; i < len(left); i++ {
						found := false
						for j := 0; j < len(right); j++ {
							if left[i].Id == right[j].Id {
								if n != nil {
									// The new merge point is after current mergepoint
									// Use the new merge point
									if n.AnyPathesContainNode(left[i]) {
										n = left[i]
									}
								} else {
									n = left[i]
								}
								found = true
								break
							}
						}
						if found {
							break
						}
					}

				}
			}
		}

	}

	return n
}

// LastNode - returns the last node in a mesh
// The last node is the last element of the likeliest path
func (m *Mesh) LastNode() *MeshNode {

	l := m.Start.likelyPath()
	return l[len(l)-1]
}

// convertDot - Convert the mesh to a dot and return as string
func (m *Mesh) convertDot(simple bool) string {
	g := dot.NewGraph(dot.Directed)

	var start = g.Node("Start")
	for i := range m.Start.Next {
		str := ""
		str += m.Start.Next[i].TLE.String()
		str += "\n"
		str += m.Start.Next[i].Hash
		str += "\n"

		for u := range m.Start.Next[i].FirmwareOptions {
			for k, v := range m.Start.Next[i].FirmwareOptions[u] {
				str += k + "=" + strconv.FormatUint(v, 10) + " "
			}
			str += "\n"
		}
		n := g.Node(strconv.FormatInt(int64(m.Start.Next[i].Id), 10))
		if simple {
			str = strconv.FormatInt(int64(m.Start.Next[i].Id), 10)
			n.Attr("color", "#"+m.Start.Next[i].Hash[0:6])
			n.Attr("fillcolor", "#"+m.Start.Next[i].Hash[0:6])
			n.Attr("style", "filled")
		}
		n.Label(str)

		g.Edge(start, n)
	}

	for i := range m.Nodes {
		//n := g.Node(string(self.Nodes[i].Hash))
		str := ""
		str += string(m.Nodes[i].TLE.String())
		str += "\n"
		str += m.Nodes[i].Hash
		str += "\n"
		for u := range m.Nodes[i].FirmwareOptions {
			for k, v := range m.Nodes[i].FirmwareOptions[u] {
				str += k + "=" + strconv.FormatUint(v, 10) + " "
			}
			str += "\n"
		}
		n := g.Node(strconv.FormatInt(int64(m.Nodes[i].Id), 10))
		if simple {
			str = strconv.FormatInt(int64(m.Nodes[i].Id), 10)
			n.Attr("color", "#"+m.Nodes[i].Hash[0:6])
			n.Attr("fillcolor", "#"+m.Nodes[i].Hash[0:6])
			n.Attr("style", "filled")

		}
		n.Label(str)

		for j := range m.Nodes[i].Next {
			m := g.Node(strconv.FormatInt(int64(m.Nodes[i].Next[j].Id), 10))
			g.Edge(n, m)
		}
	}
	return g.String()
}

// CreateNode - Create a new node, but don't add it to the mesh
func (m *Mesh) CreateNode(nop bool) *MeshNode {
	var n MeshNode

	n.IsNoop = nop
	n.Id = m.ID
	n.FirmwareOptions = []map[string]uint64{}
	m.ID++

	return &n
}

// WriteDot - Convert mesh to dot and write it to file
func (m *Mesh) WriteDot(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(m.convertDot(true))
	if err != nil {
		return err
	}
	f.Sync()
	return nil
}

// merge b into m
func mergeNodes(b *MeshNode, m *MeshNode) {
	// marked nodes on path as used
	m.Propability++
	m.Hash = b.Hash
	m.TLE = b.TLE

	if len(b.FirmwareOptions) > 0 {
		m.FirmwareOptions = append(m.FirmwareOptions, b.FirmwareOptions[0])
	}
}

// Merge branch into mesh as long as there're equal nodes in the branches
// Returns merged count
// Returns true, nil on sucessful merge
// Returns false, branchNode (left), branchNode parent (right) once a branch point is dectected
func mergeSimple(branch *Mesh, branchStart *MeshNode, mesh *Mesh, meshStart *MeshNode) (uint, bool, *MeshNode, *MeshNode) {
	var merged uint
	if branchStart == nil {
		branchStart = &branch.Start
	}
	if meshStart == nil {
		meshStart = &mesh.Start
	}

	var left = branchStart
	var right = meshStart

	if len(left.Next) == 0 {
		// Done
		return merged, true, nil, nil
	}

	for len(left.Next) > 0 {
		var path = -1

		// traverse mesh and follow nodes equal those on branch
		for i := range right.Next {
			if right.Next[i].Hash == left.Next[0].Hash {
				path = i
				break
			}
		}
		if path == -1 {
			//log.Printf("Merge: l %s / ^r %s are not equal\n", left.Next[0].Hash, right.Hash)

			// detected a branch
			return 0, false, left.Next[0], right
		} else {
			//log.Printf("Merge: l %s / r %s are equal\n", left.Next[0].Hash, right.Next[path].Hash)
		}

		mergeNodes(left.Next[0], right.Next[path])

		right = right.Next[path]
		left = left.Next[0]
		merged++
	}

	return merged, true, nil, nil
}

func mergeGetLCS(l *MeshNode, r *MeshNode) (int, []*MeshNode, []interface{}) {
	// merge tail using LCS
	var rpath = r.FirstPath()
	var bestLcs = -1
	var bestPath []*MeshNode
	var lcsResult []interface{}

	// One up to make use of firstPath method
	var lpath = l.FirstPath()

	// do LCS for all path in the mesh
	for true {
		log.Printf("l ")
		for i := range lpath {
			log.Printf("%s ", lpath[i].Hash)
		}
		log.Printf("\nr ")
		for i := range rpath {
			log.Printf("%s ", rpath[i].Hash)
		}
		log.Printf("\n")

		lcs := lcsMeshNodes(lpath, rpath)
		i := len(lcs)
		log.Printf("lcs %v\n", lcs)

		if i > bestLcs {
			lcsResult = lcs
			bestLcs = i
			bestPath = rpath
		}
		rpath = r.NextPath(rpath)
		if rpath == nil {
			break
		}
	}

	return bestLcs, bestPath, lcsResult
}

func merge(branch *Mesh, mesh *Mesh) {
	left := &branch.Start
	right := &mesh.Start
	var leftcur uint = 0
	var leftlen uint = left.LenFirstPath()

	log.Printf("Merging...\n")

	mesh.Pathes++

	for true {
		mergedcount, done, bl, br := mergeSimple(branch, left, mesh, right)
		leftcur += mergedcount
		log.Printf("merge progress [%d/%d]\n", leftcur, leftlen)

		if done {
			return
		}

		if len(bl.Prev) == 0 {
			log.Fatal("Internal error. Left node has no previous element!")
			return
		}

		bestLcs, bestPath, lcsResult := mergeGetLCS(bl.Prev[0], br)
		if bestLcs == -1 {
			log.Fatal("Internal error. Got no LCS result!")
			return
		}
		log.Printf("bestLcs %d\n", bestLcs)
		log.Printf("lcsResult %v\n", lcsResult)

		left = bl
		right = br

		if len(lcsResult) == 0 {
			old := right
			// no match, merge everything into new branch
			for true {
				var m = mesh.CreateNode(false)
				m.Propability = 1
				mergeNodes(left, m)

				mesh.insertNode(old, m)
				old = m
				if len(left.Next) == 0 {
					log.Printf("merge progress [%d/%d]\n", leftcur, leftlen)
					return
				}
				left = left.Next[0]
				leftcur++
			}
		} else {
			old := right

			// match, merge partial tree into mesh
			var j = 0
			// 1. iterate left until LCS match adding nodes onto mesh
			// 2. iterate right until LCS match
			// 3. merge left into right
			// 4. iterate right until no LCS match
			// 5. branch into new left
			// 6. goto 1
			var LCSLeftFound = false
			for true {
				if left.Hash == lcsResult[0] {
					log.Printf("M: Found left reentry point %s\n", left.Hash)

					LCSLeftFound = true
					break
				} else {
					log.Printf("M: Creating branch node for %s\n", left.Hash)

					var m = mesh.CreateNode(false)
					m.Propability = 1
					mergeNodes(left, m)
					mesh.insertNode(old, m)
					old = m
					leftlen++
				}
				if len(left.Next) > 0 {
					left = left.Next[0]
				} else {
					break
				}
			}

			var LCSRightFound = false
			for ; j < len(bestPath); j++ {
				if bestPath[j].Hash == lcsResult[0] {
					log.Printf("M: Found right reentry point %s\n", bestPath[j].Hash)

					right = bestPath[j]
					LCSRightFound = true
					break
				}
			}
			if LCSLeftFound && LCSRightFound {
				mergeNodes(left, right)
				// old is a new node created with CreateNode
				// Merge it into mesh
				old.Next = append(old.Next, right)
				right.Prev = append(right.Prev, old)
				leftlen++
			}
		}
	}

}

// appendNode - Adds the node to end of the first branch
func (m *Mesh) appendNode(newNode *MeshNode) {
	i := &m.Start

	for len(i.Next) > 0 {
		i = i.Next[0]
	}
	i.Next = append(i.Next, newNode)
	newNode.Prev = append(newNode.Prev, i)
	m.Nodes = append(m.Nodes, newNode)
}

// insertNode - Adds the node to an existing node of the mesh
func (m *Mesh) insertNode(existNode *MeshNode, newNode *MeshNode) {
	existNode.Next = append(existNode.Next, newNode)
	newNode.Prev = append(newNode.Prev, existNode)
	m.Nodes = append(m.Nodes, newNode)
}

// MeshNodeFromTraceLogEntry - Generate mesh node ftom tracelog entry
func (m *Mesh) MeshNodeFromTraceLogEntry(tle tracelog.TraceLogEntry) (*MeshNode, error) {
	var a = m.CreateNode(false)
	a.TLE = tle

	sha := sha256.Sum256([]byte(fmt.Sprintf("%v", tle)))
	a.Hash = ""
	for _, i := range sha {
		a.Hash += fmt.Sprintf("%x", i)
	}
	//log.Printf("node hash %s\n", a.Hash)
	m.Nodes = append(m.Nodes, a)

	return a, nil
}

// InsertTraceLogIntoMesh - Inserts a complete tracelog into the mesh
// It first creates a new branch and then merges the branch into the mesh using LCS
// Every created node on the branch gets assigned a FirmwareOptions slice
// On merge FirmwareOptions slices are also merged
func (m *Mesh) InsertTraceLogIntoMesh(tles []tracelog.TraceLogEntry, FirmwareOptions map[string]uint64) error {
	var b = Mesh{Start: MeshNode{Id: 0}, ID: 1}

	// Create a branch
	for i := range tles {
		n, err := b.MeshNodeFromTraceLogEntry(tles[i])
		if err != nil {
			return err
		}
		// make a deep copy
		newmap := map[string]uint64{}
		for k, v := range FirmwareOptions {
			newmap[k] = v
		}
		n.FirmwareOptions = []map[string]uint64{newmap}

		b.appendNode(n)
	}
	merge(&b, m)

	return nil
}

// Compares two unsorted slices if those have equal contents
func equalSlice(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}

	for _, v1 := range a {
		found := false
		for _, v2 := range b {
			if v1 == v2 {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}

// Compares two unsorted slices if those have equal contents
func equalMap(a, b map[string]uint64) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v1 := range a {
		if b[k] != v1 {
			return false
		}
	}
	return true
}

// Compares two unsorted slices if those have equal contents
func intInSlice(a uint64, b []uint64) bool {
	for _, v1 := range b {
		if v1 == a {
			return true
		}
	}
	return false
}

func (self *MeshNode) containsEqualFirmwareOption(FirmwareOption map[string]uint64) bool {
	for i := range self.FirmwareOptions {
		if equalMap(FirmwareOption, self.FirmwareOptions[i]) {
			return true
		}
	}
	return false
}

func deepCopyMap(c map[string]uint64) map[string]uint64 {
	new := map[string]uint64{}
	for k, v := range c {
		new[k] = v
	}
	return new
}

// OptimiseNodeByRemovingFirmwareOptions - If a node has all options a FirmwareOption can have, remove it as it doesn't depend on the config
func (mn *MeshNode) OptimiseNodeByRemovingFirmwareOptions(allFirmwareOptions map[string][]uint64) error {
	for k := range allFirmwareOptions {
		for i := range mn.FirmwareOptions {
			copy := deepCopyMap(mn.FirmwareOptions[i])

			foundAllFirmwareOptions := true
			for _, v := range allFirmwareOptions[k] {
				copy[k] = v

				// search
				if !mn.containsEqualFirmwareOption(copy) {
					foundAllFirmwareOptions = false
					break
				}
			}
			if !foundAllFirmwareOptions {
				continue
			}
			for _, v := range allFirmwareOptions[k] {
				copy[k] = v

				// search
				for j := range mn.FirmwareOptions {
					if equalMap(copy, mn.FirmwareOptions[j]) {
						delete(mn.FirmwareOptions[j], k)
					}
				}
			}
		}
	}

	// Remove duplicated lines
	for true {
		idx := -1
		for i := 0; i < len(mn.FirmwareOptions)-1; i++ {
			for j := i + 1; j < len(mn.FirmwareOptions); j++ {
				if equalMap(mn.FirmwareOptions[i], mn.FirmwareOptions[j]) {
					idx = j
					break
				}
			}
			if idx >= 0 {
				copy(mn.FirmwareOptions[idx:], mn.FirmwareOptions[idx+1:])
				mn.FirmwareOptions = mn.FirmwareOptions[:len(mn.FirmwareOptions)-1]
				break
			}
		}
		if idx < 0 {
			break
		}
	}

	return nil
}

// OptimiseMeshByRemovingFirmwareOptions - If a node has all options a FirmwareOption can have, remove it as it doesn't depend on the config
func (m *Mesh) OptimiseMeshByRemovingFirmwareOptions(allFirmwareOptions map[string][]uint64) error {
	for i := range m.Nodes {
		m.Nodes[i].OptimiseNodeByRemovingFirmwareOptions(allFirmwareOptions)
	}

	return nil
}

// MergeMeshNodesAndUnlink - merge b into m
func (m *Mesh) MergeMeshNodesAndUnlink(b *MeshNode, keep *MeshNode) {
	mergeNodes(b, keep)

	// Merge Next pointer into m
	for _, i := range b.Prev {
		found := false
		for _, j := range keep.Prev {
			if i.Id == j.Id {
				found = true
				break
			}
		}
		if !found {
			keep.Prev = append(keep.Prev, i)
			i.Next = append(i.Next, keep)
		}
	}

	// Remove Next pointer to b
	for _, i := range b.Prev {
		var nodes []*MeshNode

		for _, j := range i.Next {
			if b.Id == j.Id {
				continue
			}
			nodes = append(nodes, j)
		}
		i.Next = nodes
	}

	// Merge Next pointer into m
	for _, i := range b.Next {
		found := false
		for _, j := range keep.Next {
			if i.Id == j.Id {
				found = true
				break
			}
		}
		if !found {
			keep.Next = append(keep.Next, i)
			i.Prev = append(i.Prev, keep)
		}
	}

	// Remove Prev pointer to b
	for _, i := range b.Next {
		var nodes []*MeshNode

		for _, j := range i.Prev {
			if b.Id == j.Id {
				continue
			}
			nodes = append(nodes, j)
		}
		i.Prev = nodes
	}

	// Unlink node from mesh
	var nodes []*MeshNode
	for _, j := range m.Nodes {
		if b.Id == j.Id {
			continue
		}
		nodes = append(nodes, j)
	}
	m.Nodes = nodes
}

// OptimiseMeshByRemovingNodes - Try to remove duplicated nodes.
// The branch merging only working for all pathes below the branch point
// Start at the end of the mesh and merge duplicated nodes into one
func (m *Mesh) OptimiseMeshByRemovingNodes() error {
	log.Printf("Optimising mesh by removing nodes...\n")

	iteration := 0
	for true {
		log.Printf("Analysing tree, iteration %d\n", iteration)

		found := false
		for p := m.Start.FirstPath(); p != nil; p = m.Start.NextPath(p) {
			for i := range p {
				j := len(p) - i - 1
				if len(p[j].Prev) > 1 {

					for n := range p[j].Prev {
						if n == 0 {
							continue
						}
						if p[j].Prev[n-1].Hash == p[j].Prev[n].Hash && len(p[j].Prev[n-1].Next) == 1 && len(p[j].Prev[n].Next) == 1 {
							log.Printf("Removing node Id %s\n", strconv.FormatInt(int64(p[j].Prev[n-1].Id), 10))
							m.MergeMeshNodesAndUnlink(p[j].Prev[n-1], p[j].Prev[n])
							found = true
							break
						}
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			break
		}
		iteration++
	}

	return nil
}

// OptimiseMeshByAddingNops - Try to insert nops to reduce the number of edges
// This leads to prettier code generation
func (m *Mesh) OptimiseMeshByAddingNops() error {
	log.Printf("Optimising mesh by iserting noop nodes...\n")

	iteration := 0
	for true {
		log.Printf("Analysing tree, iteration %d\n", iteration)

		var nodes []*MeshNode
		for n1 := range m.Nodes {
			if len(m.Nodes[n1].Prev) < 2 {
				continue
			}
			if len(m.Nodes[n1].Prev) > 0 && m.Nodes[n1].Prev[0].IsNoop {
				continue
			}
			if m.Nodes[n1].IsNoop {
				continue
			}
			for n2 := range m.Nodes {
				// comparing the same node, skip
				if m.Nodes[n1].Id == m.Nodes[n2].Id {
					continue
				}
				if m.Nodes[n2].IsNoop {
					continue
				}
				if len(m.Nodes[n2].Prev) > 0 && m.Nodes[n2].Prev[0].IsNoop {
					continue
				}
				if m.Nodes[n1].comparePrev(m.Nodes[n2]) {
					log.Printf("n1\n")

					for i := range m.Nodes[n1].Prev {
						log.Printf("Prev %s\n", strconv.FormatInt(int64(m.Nodes[n1].Prev[i].Id), 10))
					}
					log.Printf("n2\n")
					for i := range m.Nodes[n2].Prev {
						log.Printf("Prev %s\n", strconv.FormatInt(int64(m.Nodes[n2].Prev[i].Id), 10))
					}
					log.Printf("Nodes %s and %s have the same prev\n", strconv.FormatInt(int64(m.Nodes[n1].Id), 10), strconv.FormatInt(int64(m.Nodes[n2].Id), 10))
					// only add n1, n2 will be added later
					nodes = append(nodes, m.Nodes[n1])
				}
			}
		}

		if len(nodes) == 0 {
			break
		}

		/* Create dummy instruction */
		noop := m.CreateNode(true)
		noop.Prev = nodes[0].Prev
		noop.Next = nodes
		noop.Hash = fmt.Sprintf("ffffff noop id=%d", noop.Id)

		m.Nodes = append(m.Nodes, noop)

		// Insert the single noop
		for j := range nodes[0].Prev {
			nodes[0].Prev[j].Next = append(nodes[0].Prev[j].Next, noop)
		}

		/* Unlink old prev / next */
		for i := range nodes {
			for j := range nodes[i].Prev {
				// remove n from n.Prev[i].Next
				var newNext []*MeshNode
				for x := range nodes[i].Prev[j].Next {
					if nodes[i].Prev[j].Next[x].Id != nodes[i].Id {
						newNext = append(newNext, nodes[i].Prev[j].Next[x])
					}
				}
				nodes[i].Prev[j].Next = newNext
			}
		}

		// Now insert the single noop
		for i := range nodes {
			nodes[i].Prev = []*MeshNode{noop}
		}
		iteration++
	}

	return nil
}
