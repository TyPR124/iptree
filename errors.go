package iptree

import "errors"

//ErrNotImplimented should no longer be used
//var ErrNotImplimented = errors.New("Not Implimented Yet")

//ErrWrongIPLength indicates you passed an IPNet value with the wrong length of IP address (maybe you passed IPv4 into a v6 tree, or vice versa?)
var ErrWrongIPLength = errors.New("IP length does not match root's IP length")

//ErrNotFound indicates the requested element was not found in the tree
var ErrNotFound = errors.New("Could not find element")

//ErrInvalidData is not currently used
//var ErrInvalidData = errors.New("Invalid data")

//ErrNewRoot indicates an insertion caused a new root element to be created
type ErrNewRoot struct {
	NewRoot Root
}

func (ErrNewRoot) Error() string {
	return "Insertion caused new root"
}

//ErrRemovedRoot is an error type indicating the root was removed. NewRoots contains all children of the previous root.
type ErrRemovedRoot struct {
	NewRoots []Root
}

func (ErrRemovedRoot) Error() string {
	return "Root element removed"
}
