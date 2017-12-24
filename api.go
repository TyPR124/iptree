package iptree

import (
	"net"
)

//NewDefaultRoot returns a new Root element of all zeros (ie, 0.0.0.0 if lenght is 4)
func NewDefaultRoot(length int, rootValue interface{}) Root {
	b := make([]byte, length)
	ipnet := net.IPNet{
		IP:   net.IP(b),
		Mask: net.IPMask(b),
	}
	return makeNode(ipnet, rootValue, nil)
}

//NewRoot returns a new Root with
func NewRoot(ipnet net.IPNet, rootValue interface{}) Root {
	return makeNode(ipnet, rootValue, nil)
}

//TraversalFunc is a type passed to Root.Traverse().
//It must accept an IPNet, generic value, and a distance value.
//Distance indicates the distance from root. Root always has distance 0.
type TraversalFunc func(ipnet net.IPNet, value interface{}, distance int)

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
	//distance indicates the distances from root. Root has distance 0.
	Traverse(TraversalFunc)
}
