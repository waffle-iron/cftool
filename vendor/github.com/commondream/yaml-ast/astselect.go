package yamlast

import (
	"regexp"
	"strconv"
	"strings"
)

const selectorRegex = `(([^\.\"\[\]\s]+)|\"([^\"]*)\"|\[([0-9]+)\])\.?`

// SelectNode uses a selector string to find and return the node described.
// If no node is described, nil is returned.
func SelectNode(node *Node, selector string) *Node {
	if node == nil {
		return nil
	}

	selector = strings.TrimSpace(selector)

	r, _ := regexp.Compile(selectorRegex)
	matches := r.FindAllStringIndex(selector, -1)

	if len(matches) == 0 {
		return nil
	}

	// if this is a document node, grab its first child
	current := node
	if node.Kind == DocumentNode {
		if len(node.Children) > 0 {
			current = node.Children[0]
		} else {
			return nil
		}
	}

	index := 0
	for index < len(selector) {
		loc := r.FindStringSubmatchIndex(selector[index:])
		if loc == nil || loc[0] != 0 {
			return nil
		}

		key := ""
		stringKey := ""
		arrKey := ""
		if loc[4] != -1 {
			key = selector[index+loc[4] : index+loc[5]]
		}

		if loc[6] != -1 {
			stringKey = selector[index+loc[6] : index+loc[7]]
		}

		if loc[8] != -1 {
			arrKey = selector[index+loc[8] : index+loc[9]]
		}

		if key == "" && stringKey != "" {
			key = stringKey
		}

		if key != "" && current.Kind == MappingNode {
			for i := 0; i < len(current.Children); i += 2 {
				if current.Children[i].Value == key {
					current = current.Children[i+1]
					break
				}
			}
		} else if arrKey != "" && current.Kind == SequenceNode {
			i, err := strconv.Atoi(arrKey)
			if err != nil || i < 0 || i >= len(current.Children) {
				return nil
			}
			current = current.Children[i]
		} else {
			return nil
		}

		index += loc[1]
	}
	return current
}
