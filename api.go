package iptree

import (
	"io"
	"net"
)

//NewDefaultRoot returns a new Root element of all zeros (ie, 0.0.0.0 if lenght is 4)
func NewDefaultRoot(length int, rootValue interface{}) Root {
	b := make([]byte, length)
	ipnet := net.IPNet{ //IP and Mask of all 0
		IP:   net.IP(b),
		Mask: net.IPMask(b),
	}
	return makeNode(ipnet, rootValue, nil)
}

//NewRoot returns a new Root with
func NewRoot(ipnet net.IPNet, rootValue interface{}) Root {
	return makeNode(ipnet, rootValue, nil)
}

//Traverser is a type passed to Root.Traverse().
//It must accept an IPNet, a generic value, and a distance value.
//Distance indicates the distance from root. Root always has distance 0.
//If any (non-nil) error is returned, Root.Traverse() will terminate.
type Traverser func(ipnet net.IPNet, value interface{}, distance int) error

type ValueSerializer func(value interface{}) ([]byte, error)

type ValueDeserializer func(vbytes []byte) (value interface{}, e error)

//Root is the root element of the tree contains functions for manipulating the tree
type Root interface {
	//Find an element at IPNet.
	//If allowSupernet is false, the function will only return an exact IPNet match
	//If allowSupernet is true, the function will return the best match
	//If no suitable match if found, returns nil and ErrNotFound
	Find(ipnet net.IPNet, allowSupernet bool) (interface{}, error)

	//Insert inserts or overwrites an element into the tree
	Insert(net.IPNet, interface{}) error

	//Remove deletes an element at IPNet
	Remove(net.IPNet) error

	//Traverse calls the passed-in function for every element.
	//If TraversalFunc returns an error at any time, execution ends and the error is returned
	Traverse(Traverser) error

	//GetIPLength returns the length of IP Address expected
	GetIPLength() int

	//Get number of nodes in the tree
	Count() int
}

func Serialize(root Root, out io.Writer, serializer ValueSerializer) error {
	return serialize(root, out, serializer)
}

func Deserialize(in io.Reader, deserializer ValueDeserializer) (Root, error) {
	return deserialize(in, deserializer)
}
