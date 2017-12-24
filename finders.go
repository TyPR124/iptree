package iptree

import "net"

//Same as Find, but returns a node instead of a value
//Also takes caller and index arguments which should only be used when calling recursively
//Also returns the parent of the node found, if there is one
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

//Same as Find, but returns a node instead of a value
//Returns the parent to insert in, at what index (-1 if the parent is actually the node to overwrite), and how many children to move
//checkNext and amChild are for recusive calls only
func (n *node) findForInsertion(mark net.IPNet) (parent *node, atIndex, numChildren int, checkNext, amChild bool, err error) {
	maskdiff := compareMask(n.Mask, mark.Mask)
	ipdiff := compareIP(n.IP, mark.IP)

	//Check next child only if we are less than mark

	if maskdiff == 0 { //Mask is the same...
		if ipdiff == 0 { //And netw addr is the same, therefore I am an exact match
			return n, -1, 0, false, false, nil
		}
		//And netw addr is different, therefore we do not contain it so it cannot be found
		//Check next child only if we are less than mark
		checkNext = (ipdiff < 0)
		return nil, -1, 0, checkNext, false, ErrNotFound
	}

	checkNext = (ipdiff < 0)

	if maskdiff > 0 { //My mask is more specific, therefore it might contain me
		return nil, -1, 0, checkNext, mark.Contains(n.IP), ErrNotFound
	}

	//Mark's mask is more specific...
	amChild = false

	if !n.Contains(mark.IP) { //but I do not contain it
		return nil, -1, 0, checkNext, false, ErrNotFound
	}

	//I contain it, therefore I am a supernet...
	//A child might contain it, or it may contain a set of my children

	foundChild := false
	chInd := 0
	nCh := 0

	for i, c := range n.children { //I might have a child who can find it
		parent, atIndex, numChildren, checkNext, isChild, _ := c.findForInsertion(mark)

		if parent != nil { //The child found it
			return parent, atIndex, numChildren, false, false, nil
		}

		if isChild {
			nCh++
			if !foundChild {
				chInd = i
				foundChild = true
			}
			continue //Get all consecutive children
		}

		if !checkNext { //If this child is too high, insert at this index
			if foundChild {
				return n, chInd, nCh, false, false, nil
			}
			return n, i, 0, false, false, nil
		}
	}

	//I have no children, or no children are higher than it
	//If mark has children, then return that info
	if foundChild {
		return n, chInd, nCh, false, false, nil
	}
	return n, len(n.children), 0, false, false, nil
}

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
