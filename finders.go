package iptree

import "net"

//Recursively look for a node to be removed.
//Must be an exact match.
//caller and index are intended for recursive calls only.
//Returns the node to be removed, the parent node, and the index where
//the node to be removed exists within the parent.children slice.
//If n is the node to be removed, caller and index are passed through and returned as parent and childIndex
func (n *node) findForRemoval(mark net.IPNet, caller *node, index int) (vnode *node, parent *node, childIndex int, err error) {
	maskdiff := compareMask(n.Mask, mark.Mask)

	if maskdiff == 0 { //Mask is the same...
		if sameIP(n.IP, mark.IP) { //And netw addr is the same, therefore I am the one
			return n, caller, index, nil
		}
		//And netw addr is different, therefore we do not contain it so it cannot be found
		return nil, nil, -1, ErrNotFound
	}

	if maskdiff > 0 { //My mask is more specific, therefore I am not it and do not contain it
		return nil, nil, -1, ErrNotFound
	}

	//Mark's mask is more specific...
	if !n.Contains(mark.IP) { //but I do not contain it
		return nil, nil, -1, ErrNotFound
	}

	//I contain it, therefore I am a supernet...
	for i, c := range n.children { //and I might have a child who can find it
		vnode, parent, childIndex, err = c.findForRemoval(mark, n, i)

		if vnode != nil { //The child found it
			return
		}
	}

	return nil, nil, -1, ErrNotFound
}

//Recursively look for a place to insert a node.
//Returns:
// parent: the parent in which to insert the new node
// atIndex: where to insert the node
// numChildren: the number of children currently belonging to parent that must be moved to the new node
// checkNext: for recursive calls only. Whether we should check the next child
// amChild: true if n is a child of mark
// err: error
func (n *node) findForInsertion(mark net.IPNet) (parent *node, atIndex, numChildren int, checkNext, amChild bool, err error) {
	maskdiff := compareMask(n.Mask, mark.Mask)
	ipdiff := compareIP(n.IP, mark.IP)

	//Return values
	parent = nil
	atIndex = -1
	numChildren = 0
	checkNext = ipdiff < 0
	amChild = false
	err = nil

	//Check next child only if we are less than mark

	if maskdiff == 0 { //Mask is the same...
		if ipdiff == 0 { //And netw addr is the same, therefore I am an exact match
			parent = n
			return
		}
		//And netw addr is different, therefore we do not contain it so it cannot be found
		//Check next child only if we are less than mark
		err = ErrNotFound
		return
	}

	checkNext = (ipdiff < 0)

	if maskdiff > 0 { //My mask is more specific, therefore it might contain me
		amChild = mark.Contains(n.IP)
		err = ErrNotFound
		return
	}

	//Mark's mask is more specific...

	if !n.Contains(mark.IP) { //but I do not contain it
		err = ErrNotFound
		return
	}

	//I contain it, therefore I am a supernet...
	//A child might contain it, or it may contain a set of my children

	checkNext = false

	foundChild := false
	chInd := 0
	nCh := 0

	for i, c := range n.children { //I might have a child who can find it
		var doCheckNext, isChild bool
		//Avoid using := on recursive call, it causes parent, atIndex, numChildren to become new variables, shadowing the actual return variables
		parent, atIndex, numChildren, doCheckNext, isChild, _ = c.findForInsertion(mark)

		if parent != nil { //The child found it
			return
		}

		if isChild {
			nCh++
			if !foundChild {
				chInd = i
				foundChild = true
			}
			continue //Get all consecutive children
		}

		if !doCheckNext { //If this child is too high, insert at this index
			parent = n
			if foundChild {
				atIndex = chInd
				numChildren = nCh
				return
			}
			atIndex = i
			return
		}
	}

	if foundChild { //If a child was found but we fell through the loop before returning
		//then return now
		parent = n
		atIndex = chInd
		numChildren = nCh
		return
	}

	//I have no children, or no children are higher than it, so insert it last
	parent = n
	atIndex = len(n.children)
	return
}

//Recursively find a node.
func (n *node) findNode(mark net.IPNet, allowSupernet bool) (vnode *node, err error) {
	maskdiff := compareMask(n.Mask, mark.Mask)

	if maskdiff == 0 { //Mask is the same...
		if sameIP(n.IP, mark.IP) { //And netw addr is the same, therefore I am the one
			return n, nil
		}
		//And netw addr is different, therefore we do not contain it so it cannot be found
		return nil, ErrNotFound
	}

	if maskdiff > 0 { //My mask is more specific, therefore I am not it and do not contain it
		return nil, ErrNotFound
	}

	//Mark's mask is more specific...
	if !n.Contains(mark.IP) { //but I do not contain it
		return nil, ErrNotFound
	}

	//I contain it, therefore I am a supernet...
	for _, c := range n.children { //and I might have a child who can find it
		vnode, err = c.findNode(mark, allowSupernet)
		if vnode != nil { //The child found it
			return
		}
	}

	//No children found it, so return myself if allowed
	if allowSupernet {
		return n, nil
	}

	return nil, ErrNotFound
}
