package merkledag

import (
	"encoding/json"
	"hash"
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

func Add(store KVStore, node Node, h hash.Hash) []byte {
	switch node.Type() {
	case DIR:
		file := node.(File)
		tmp := StoreFile(store, file, h)
		jsonMarshal, _ := json.Marshal(tmp)
		hash := calculateHash(jsonMarshal, h)
		return hash
	case FILE:
		dir := node.(Dir)
		tmp := StoreDir(store, dir, h)
		jsonMarshal, _ := json.Marshal(tmp)
		hash := calculateHash(jsonMarshal, h)
		return hash
	}
	panic("unknown node")
}

func calculateHash(data []byte, h hash.Hash) []byte {
	h.Reset()
	hash := h.Sum(data)
	h.Reset()
	return hash
}

func StoreFile(store KVStore, file File, h hash.Hash) *Object {
	data := file.Bytes()
	blob := Object{Data: data, Links: nil}
	jsonMarshal, _ := json.Marshal(blob)
	hash := calculateHash(jsonMarshal, h)
	store.Put(hash, data)
	return &blob
}

func StoreDir(store KVStore, dir Dir, h hash.Hash) *Object {
	it := dir.It()
	treeObject := &Object{}
	for it.Next() {
		n := it.Node() //当前目录下的node
		switch n.Type() {
		case FILE:
			file := n.(File)
			tmp := StoreFile(store, file, h)
			jsonMarshal, _ := json.Marshal(tmp)
			hash := calculateHash(jsonMarshal, h)
			treeObject.Links = append(treeObject.Links, Link{
				Hash: hash,
				Size: int(file.Size()),
				Name: file.Name(),
			})
			typeName := "link"
			if tmp.Links == nil {
				typeName = "blob"
			}
			treeObject.Data = append(treeObject.Data, []byte(typeName)...)
		case DIR:
			dir := n.(Dir)
			tmp := StoreDir(store, dir, h)
			jsonMarshal, _ := json.Marshal(tmp)
			hash := calculateHash(jsonMarshal, h)
			treeObject.Links = append(treeObject.Links, Link{
				Hash: hash,
				Size: int(dir.Size()),
				Name: dir.Name(),
			})
			typeName := "tree"
			treeObject.Data = append(treeObject.Data, []byte(typeName)...)
		}
	}
	jsonMarshal, _ := json.Marshal(treeObject)
	hash := calculateHash(jsonMarshal, h)
	store.Put(hash, jsonMarshal)
	return treeObject
}
