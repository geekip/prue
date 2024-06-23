package prue

import (
	"strings"
)

type trie struct {
	root *trieNode
}

type trieNode struct {
	middlewares   []Handler
	handlers      []Handler
	methods       map[string]Handler
	paramKey      string
	params        map[string]string
	children      map[string]*trieNode
	paramChild    *trieNode
	wildcardChild *trieNode
	isEnd         bool
}

const (
	paramPrefix    = ":"
	wildcardPrefix = "*"
)

func newTrie() *trie {
	return &trie{root: newTrieNode()}
}

func newTrieNode() *trieNode {
	return &trieNode{
		children: make(map[string]*trieNode),
		methods:  make(map[string]Handler),
		params:   make(map[string]string),
	}
}

func (t *trie) add(method, pattern string, handler Handler, middlewares []Handler) {
	node := t.root
	segments := strings.Split(pattern, "/")
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		if strings.HasPrefix(segment, paramPrefix) {
			if node.paramChild == nil {
				node.paramChild = newTrieNode()
			}
			node = node.paramChild
			node.paramKey = segment[len(paramPrefix):]
		} else if segment == wildcardPrefix {
			if node.wildcardChild == nil {
				node.wildcardChild = newTrieNode()
			}
			node = node.wildcardChild
			node.paramKey = wildcardPrefix
			break
		} else {
			if _, exists := node.children[segment]; !exists {
				node.children[segment] = newTrieNode()
			}
			node = node.children[segment]
		}
	}
	node.methods[method] = handler
	node.middlewares = append(node.middlewares, middlewares...)
	node.isEnd = true
}

func (t *trie) find(method, pattern string) *trieNode {
	node := t.root
	segments := strings.Split(pattern, "/")
	params := make(map[string]string)
	for i, segment := range segments {
		if segment == "" {
			continue
		}
		if child, ok := node.children[segment]; ok {
			node = child
			continue
		}
		if node.paramChild != nil {
			node = node.paramChild
			params[node.paramKey] = segment
			continue
		}
		if node.wildcardChild != nil {
			node = node.wildcardChild
			params[node.paramKey] = strings.Join(segments[i:], "/")
			break
		}
		return nil
	}

	if node.isEnd {
		handler := node.methods[method]
		if handler == nil {
			handler = node.methods[wildcardPrefix]
		}
		node.handlers = append(node.middlewares, handler)
		node.params = params
		return node
	}
	return nil
}
