package core

import (
	"FileBackup/internal/log"
	"bytes"
)

type Diff struct {
	A    *DirNode
	B    *DirNode
	Diff *DirNode
}

func compareHelper(a *DirNode, b *DirNode, d *DirNode) {
	for fileName, sumA := range a.File {
		sumB, exists := b.File[fileName]
		if !exists {
			continue
		}
		if !bytes.Equal(sumA, sumB) {
			d.File[fileName] = nil
		}
		delete(a.File, fileName)
		delete(b.File, fileName)
	}
	for dirPath, aNode := range a.Dir {
		bNode, exists := b.Dir[dirPath]
		if !exists {
			continue
		}
		d.Dir[dirPath] = initNode()
		compareHelper(aNode, bNode, d.Dir[dirPath])
		if len(aNode.Dir) == 0 && len(aNode.File) == 0 {
			delete(a.Dir, dirPath)
		}
		if len(bNode.Dir) == 0 && len(bNode.File) == 0 {
			delete(b.Dir, dirPath)
		}
		if len(d.Dir[dirPath].Dir) == 0 && len(d.Dir[dirPath].File) == 0 {
			delete(d.Dir, dirPath)
		}
	}
}

func Compare(a *Backup, b *Backup) (diff *Diff, err error) {
	_a := a.GetNodeAtPath(a.Path)
	_b := b.GetNodeAtPath(b.Path)

	// If one of _a or _b is nil, the other contains the exclusive difference.
	// This is means d is nil as there cannot be common nodes.
	if _a == nil || _b == nil {
		return
	}

	diff = &Diff{Diff: initNode()}

	// Easy, dirty way of deep copying
	if diff.A, err = deepCopyDirNode(_a); err != nil {
		log.Debug("deep copy error:", err)
		return
	}
	if diff.B, err = deepCopyDirNode(_b); err != nil {
		log.Debug("deep copy error:", err)
		return
	}

	compareHelper(diff.A, diff.B, diff.Diff)

	return
}
