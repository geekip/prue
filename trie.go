package prue

import (
	"regexp"
	"strings"
)

type (
	node struct {
		key         string
		pattern     string
		handler     Handler
		middlewares []Handler
		methods     map[string]Handler
		children    map[string]*node
		params      map[string]string
		paramName   string
		paramNode   *node
		regex       *regexp.Regexp
		isEnd       bool
	}
	reMaps map[string]*regexp.Regexp
)

var (
	regexCache     reMaps = make(reMaps)
	regexpPrefix   string = ":"
	wildcardPrefix string = "*"
	paramPrefix    string = "{"
	paramSuffix    string = "}"
)

// creates and caches a regular expression pattern
func compileRegex(pattern string) *regexp.Regexp {
	if re, exists := regexCache[pattern]; exists {
		return re
	}
	re := regexp.MustCompile("^" + pattern + "$")
	regexCache[pattern] = re
	return re
}

func findCommonPrefixLength(str1, str2 string) int {
	length := len(str1)
	if len(str2) < length {
		length = len(str2)
	}
	for i := 0; i < length; i++ {
		if str1[i] != str2[i] {
			return i
		}
	}
	return length
}

// creates node
func newNode(key string) *node {
	return &node{
		key:      key,
		children: make(map[string]*node),
		methods:  make(map[string]Handler),
		params:   make(map[string]string),
	}
}

// insert node
func (n *node) add(method, pattern string, handler Handler, middlewares []Handler) *node {
	if method == "" || pattern == "" || handler == nil {
		return nil
	}
	segments := strings.Split(pattern, "/")
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		// insert parameter node
		if strings.HasPrefix(segment, paramPrefix) && strings.HasSuffix(segment, paramSuffix) {
			if n.paramNode == nil {
				n.paramNode = newNode(segment)
			}
			param := segment[1 : len(segment)-1]
			parts := strings.SplitN(param, regexpPrefix, 2)
			n.paramNode.paramName = parts[0]
			if len(parts) == 2 {
				n.paramNode.regex = compileRegex(parts[1])
			}
			n = n.paramNode
		} else {
			// insert static node
			n = n.addStaticSegment(n, segment)
		}
	}
	n.isEnd = true
	n.pattern = pattern
	n.methods[method] = handler
	n.middlewares = append(n.middlewares, middlewares...)
	return n
}

func (n *node) addStaticSegment(node *node, segment string) *node {
	for len(segment) > 0 {
		found := false
		for key, child := range node.children {
			commonLength := findCommonPrefixLength(segment, key)
			if commonLength == 0 {
				continue
			}

			commonPrefix := segment[:commonLength]
			remainingSegment := segment[commonLength:]
			remainingKey := key[commonLength:]

			if len(remainingKey) > 0 {
				commonNode := newNode(commonPrefix)
				node.children[commonPrefix] = commonNode
				commonNode.children[remainingKey] = child
				delete(node.children, key)
				child.key = remainingKey

				if len(remainingSegment) > 0 {
					commonNode.children[remainingSegment] = newNode(remainingSegment)
					return commonNode.children[remainingSegment]
				}
				return commonNode
			}
			node = child
			segment = remainingSegment
			found = true
			break
		}

		if !found {
			node.children[segment] = newNode(segment)
			return node.children[segment]
		}
	}
	return node
}

func (n *node) matchParam(segment string, reSegments []string) (*node, string) {
	node := n.paramNode
	// match wildcard parameter
	if strings.HasPrefix(node.paramName, wildcardPrefix) {
		return node, strings.Join(reSegments, "/")
	}
	// parameter node has a regex
	if node.regex != nil && !node.regex.MatchString(segment) {
		return nil, ""
	}
	return node, segment
}

func (n *node) matchStaticSegment(segment string) *node {
	if n.children[segment] != nil {
		return n.children[segment]
	}
	for len(segment) > 0 {
		var found bool
		for key, child := range n.children {
			if strings.HasPrefix(segment, key) {
				n = child
				segment = segment[len(key):]
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return n
}

// Finds a matching route node
func (n *node) find(method, url string) *node {
	params := make(map[string]string)
	segments := strings.Split(url, "/")
	for i, segment := range segments {
		if segment == "" {
			continue
		}
		if node := n.matchStaticSegment(segment); node != nil {
			n = node
		} else if n.paramNode != nil {
			if node, param := n.matchParam(segment, segments[i:]); node != nil {
				n = node
				params[n.paramName] = param
			}
		} else {
			return nil
		}
	}

	if n.isEnd {
		// get handler
		handler := n.methods[method]
		if handler == nil {
			handler = n.methods[wildcardPrefix]
		}
		n.params = params
		n.handler = handler
		return n
	}
	return nil
}
