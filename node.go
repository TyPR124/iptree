package iptree

import "net"

//A node in the tree.
//Every node is technically a Root.
//See API documentation for info on exported functions in this file.
type node struct {
	net.IPNet
	value    interface{}
	children []*node //Pointers to children, so we don't have to move in-memory nodes on insertion, just move the pointers
}

func makeNode(ipnet net.IPNet, value interface{}, children []*node) *node {
	return &node{ipnet, value, children}
}

func (n *node) Find(ipnet net.IPNet, allowSupernet bool) (value interface{}, err error) {
	if !sameIPLen(n.IPNet, ipnet) {
		return nil, ErrWrongIPLength
	}

	vnode, err := n.findNode(ipnet, allowSupernet)
	if err != nil {
		return nil, err
	}

	return vnode.value, nil
}

func (n *node) Insert(ipnet net.IPNet, value interface{}) error {
	if !sameIPLen(n.IPNet, ipnet) {
		return ErrWrongIPLength
	}

	p, atIndex, nChildren, _, amChild, err := n.findForInsertion(ipnet)
	if err == ErrNotFound && amChild {
		//The node to be inserted can be a new root
		return ErrNewRoot{makeNode(ipnet, value, []*node{n})}
	} else if err != nil {
		return err
	}
	if atIndex == -1 { //p is the exact node, therefore overwrite value
		p.value = value
		return nil
	}

	lo := p.children[:atIndex]
	move := p.children[atIndex : atIndex+nChildren]
	hi := p.children[atIndex+nChildren:]

	newChild := makeNode(ipnet, value, move)
	p.children = make([]*node, 0, len(lo)+len(hi)+1)
	p.children = append(p.children, lo...)
	p.children = append(p.children, newChild)
	p.children = append(p.children, hi...)
	return nil
}

func (n *node) Remove(ipnet net.IPNet) error {
	if !sameIPLen(n.IPNet, ipnet) {
		return ErrWrongIPLength
	}

	rem, p, ci, err := n.findForRemoval(ipnet, nil, 0)
	if err != nil {
		return err
	}

	if p != nil {
		oldLen := len(p.children)
		if oldLen == 1 {
			p.children = nil
		} else {
			p.children = append(p.children[:ci], p.children[ci+1:]...)
		}

		return nil
	}

	//No parent was found, but no error means that the node to be removed is this node
	//In otherwords, rem must be equal to n
	if n != rem {
		panic("n != rem in Remove function")
	}

	//Every child of this node is now a root
	newRoots := make([]Root, len(n.children))
	for i, n := range n.children {
		newRoots[i] = n
	}

	return ErrRemovedRoot{newRoots}
}

func (n *node) Traverse(f Traverser) error {
	return n.traverseRecursively(f, 0)
}

//traverseRecursively is the implimentation used for Traverse
func (n *node) traverseRecursively(f Traverser, dist int) error {
	if err := f(n.IPNet, n.value, dist); err != nil {
		return err
	}
	for _, c := range n.children {
		if err := c.traverseRecursively(f, dist+1); err != nil {
			return err
		}
	}
	return nil
}

func (n *node) GetIPLength() int {
	return len(n.IP)
}

func (n *node) Count() int {
	count := 0
	for _, c := range n.children {
		count += c.Count()
	}
	return count + 1
}
