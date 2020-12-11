package main

import (
	"fmt"
	"strings"

	"github.com/breise/rstack"
)

func copyAndExpand(node interface{}) (interface{}, error) {
	// fmt.Println("============ copyAndExpand ================")
	// fmt.Printf("node: %s\n", node)
	rv, err := cpAndExp(rstack.New(), node, node)
	if err != nil {
		return nil, fmt.Errorf("cannot copyAndExpand(): %s", err)
	}
	return rv, nil
}

func cpAndExp(refsRStack *rstack.RStack, root, node interface{}) (interface{}, error) {
	var rv interface{}
	if m, isMap := node.(map[interface{}]interface{}); isMap {
		tmpMap := map[interface{}]interface{}{}
		for name, val := range m {
			// look for `$ref` as the key of a map.
			// If name is $ref, we are done with this map node.
			// Nor are we interested in filling out this tmpMap.
			// Just cpAndExp the reference and return it
			if name == `$ref` {
				// have we seen this ref already? If so, it's a cycle
				if err := detectCycle(refsRStack, val); err != nil {
					return nil, err
				}
				reference, err := resolveRef(root, val)
				if err != nil {
					return nil, fmt.Errorf("cannot resolveRef(): %s", err)
				}
				return cpAndExp(refsRStack.Push(val), root, reference)
			}
			// nameString, ok := name.(string)
			// if !ok {
			// 	return nil, fmt.Errorf("cannot cast '%v' to a string", name)
			// }
			// fmt.Printf("refsRStack: %v; key: %v, node: %v\n", refsRStack.ToSlice(), name, val)

			var err error
			tmpMap[name], err = cpAndExp(refsRStack, root, val)
			if err != nil {
				return nil, err
			}
		}
		rv = tmpMap
	} else if a, isArray := node.([]interface{}); isArray {
		tmp := make([]interface{}, len(a))
		for i, val := range a {
			var err error
			tmp[i], err = cpAndExp(refsRStack, root, val)
			if err != nil {
				return nil, err
			}
		}
		rv = tmp
	} else {
		rv = node
	}
	return rv, nil
}

func resolveRef(root, refVal interface{}) (interface{}, error) {
	valString, ok := refVal.(string)
	if !ok {
		return nil, fmt.Errorf("cannot cast $ref value (%v) to a string", refVal)
	}
	pathEls := strings.Split(valString, `/`)
	if pathEls[0] != `#` {
		return nil, fmt.Errorf("expecting first element of $ref value path to be `#`.  Got %s", pathEls[0])
	}

	return find(root, pathEls[1:])
}

func find(node interface{}, names []string) (interface{}, error) {
	if len(names) == 0 {
		return node, nil
	}
	name := names[0]
	m, isMap := node.(map[interface{}]interface{})
	if !isMap {
		return nil, fmt.Errorf("node is not a map.  It is a %T", node)
	}
	next, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("no such key: %s", name)
	}
	return find(next, names[1:])
}

func detectCycle(refsRStack *rstack.RStack, val interface{}) error {
	for _, v := range refsRStack.ToSlice() {
		if val == v {
			refsPath := refsRStack.Push(val).Join(" -> ")
			return fmt.Errorf("$ref cycle detected: %s", refsPath)
		}
	}
	return nil
}
