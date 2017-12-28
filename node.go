package iptree

import "net"

type node struct {
	net.IPNet
	value    interface{}
	children []*node //Pointers to children, so we don't have to move in-memory nodes on insertion, just move the pointers
}

func makeNode(ipnet net.IPNet, value interface{}, children []*node) *node {
	return &node{ipnet, value, children}
}

//Find an element at IPNet.
//If allowSupernet is false, the function will only return an exact IPNet match
//If allowSupernet is true, the function will return an exact match or a matching supernet
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

	p, atIndex, nChildren, _, _, err := n.findForInsertion(ipnet)
	if err == ErrNotFound && ipnet.Contains(n.IP) {
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

//Remove deletes an element at IPNet
//If ipnet happens to be the root element, then set root value to nil and return ErrRootRemoved
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

	//If we got here without return, and n != rem, then my logic is bad
	if n != rem {
		panic("n != rem in Remove function")
	}

	return ErrRemovedRoot{n.children}
}

func (n *node) Traverse(f Traverser) error {
	return n.traverseRecursively(f, 0)
}

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
