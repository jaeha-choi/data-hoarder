package core

import (
	"FileBackup/internal/log"
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxDirDepth int16 = 500

type Backup struct {
	Path string
	Head *DirNode
}

type DirNode struct {
	File map[string][]byte
	Dir  map[string]*DirNode
}

func Initialize() *Backup {
	return &Backup{Path: "", Head: nil}
}

func initNode() *DirNode {
	return &DirNode{
		File: make(map[string][]byte),
		Dir:  make(map[string]*DirNode),
	}
}

func ReadIndex(filename string) (b *Backup, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			log.Error("error:", err.Error())
		}
	}(file)

	// TODO: HMAC

	err = gob.NewDecoder(file).Decode(&b)
	return
}

func (b *Backup) WriteIndex(filename string) (err error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			log.Error("error:", err.Error())
		}
	}(file)

	// TODO: HMAC

	return gob.NewEncoder(file).Encode(&b)
}

func deepCopyDirNode(node *DirNode) (copy *DirNode, err error) {
	var buffer bytes.Buffer
	if err = gob.NewEncoder(&buffer).Encode(&node); err != nil {
		return nil, err
	}

	if err = gob.NewDecoder(&buffer).Decode(&copy); err != nil {
		return nil, err
	}
	return
}

func readLocalHelper(path *string, curr *DirNode, currDepth int16) (err error) {
	if curr == nil {
		return
	}
	entries, err := os.ReadDir(*path)
	if err != nil {
		return
	}
	var newNode *DirNode
	var newPath string
	for _, entry := range entries {
		newNode = initNode()
		newPath = *path + string(os.PathSeparator) + entry.Name()
		if entry.IsDir() {
			curr.Dir[entry.Name()] = newNode
			if currDepth >= maxDirDepth {
				log.Error("Max recursion depth reached, deeper directories will not be indexed.")
				return
			}
			if err = readLocalHelper(&newPath, newNode, currDepth+1); err != nil {
				return err
			}
		} else {
			file, err := os.Open(newPath)
			if err != nil {
				log.Error("error:", err.Error())
				if err = file.Close(); err != nil {
					log.Error("error:", err.Error())
					return err
				}
				return err
			}

			h := md5.New()
			if _, err := io.Copy(h, file); err != nil {
				log.Error("error:", err.Error())
				if err = file.Close(); err != nil {
					log.Error("error:", err.Error())
					return err
				}
				return err
			}

			//curr.File[entry.Name()] = hex.EncodeToString(h.Sum(nil))
			curr.File[entry.Name()] = h.Sum(nil)
			if err = file.Close(); err != nil {
				log.Error("error:", err.Error())
				return err
			}
		}
	}
	return
}

func (b *Backup) ReadLocal(path string) (err error) {
	if path, err = filepath.Abs(path); err != nil {
		return
	}
	b.Path = path
	entries := strings.Split(path, string(os.PathSeparator))
	b.Head = initNode()
	curr := b.Head

	var newNode *DirNode
	for _, entry := range entries {
		newNode = initNode()
		curr.Dir[entry] = newNode
		curr = newNode
	}

	return readLocalHelper(&path, curr, 0)
}

func (b *Backup) GetNodeAtPath(path string) (node *DirNode) {
	path, err := filepath.Abs(path)
	if err != nil {
		return
	}

	entries := strings.Split(path, string(os.PathSeparator))

	var ok bool
	node = b.Head
	for _, entry := range entries {
		if node, ok = node.Dir[entry]; !ok {
			return
		}
	}

	return
}

func stringHelper(curr *DirNode, space int) (str string) {
	if curr == nil {
		return "<nil>"
	}
	spacer := ""
	for i := 0; i < space; i++ {
		spacer += "  "
	}
	for pathStr, node := range curr.Dir {
		str += spacer + pathStr + "/\n" + stringHelper(node, space+1)
	}
	for pathStr, sum := range curr.File {
		str += spacer + pathStr + ": "
		if sum != nil {
			str += hex.EncodeToString(sum)[:10]
		}
		str += "...\n"
	}
	return str
}

func (d *DirNode) String() string {
	return stringHelper(d, 0)
}

func (b *Backup) String() string {
	return stringHelper(b.Head, 0)
}
