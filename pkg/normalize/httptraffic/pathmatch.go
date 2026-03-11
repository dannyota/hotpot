package httptraffic

import "strings"

// MatchEndpoint holds the data needed for path matching.
type MatchEndpoint struct {
	ID          string
	URIPattern  string
	Methods     []string
	Service     string
	AccessLevel string
}

// trieNode is a node in the segment-based path trie.
type trieNode struct {
	children map[string]*trieNode
	wildcard *trieNode // "*" single-segment wildcard
	catchAll bool      // trailing "*" catch-all
	endpoint *MatchEndpoint
}

// PathMatcher uses a segment-based trie for URI matching.
type PathMatcher struct {
	root *trieNode
}

// NewPathMatcher builds a trie from the given endpoints.
func NewPathMatcher(endpoints []MatchEndpoint) *PathMatcher {
	root := &trieNode{children: make(map[string]*trieNode)}

	for i := range endpoints {
		ep := &endpoints[i]
		segments := splitPath(ep.URIPattern)

		node := root
		for j, seg := range segments {
			if seg == "*" && j == len(segments)-1 {
				// Trailing wildcard: catch-all.
				node.catchAll = true
				node.endpoint = ep
				break
			}
			if seg == "*" {
				// Single-segment wildcard.
				if node.wildcard == nil {
					node.wildcard = &trieNode{children: make(map[string]*trieNode)}
				}
				node = node.wildcard
			} else {
				child, ok := node.children[seg]
				if !ok {
					child = &trieNode{children: make(map[string]*trieNode)}
					node.children[seg] = child
				}
				node = child
			}
		}
		if node.endpoint == nil {
			node.endpoint = ep
		}
	}

	return &PathMatcher{root: root}
}

// Match finds the best-matching endpoint for the given URI.
// Returns nil if no match found.
func (pm *PathMatcher) Match(uri string) *MatchEndpoint {
	segments := splitPath(uri)
	return pm.matchNode(pm.root, segments, 0)
}

func (pm *PathMatcher) matchNode(node *trieNode, segments []string, depth int) *MatchEndpoint {
	if depth == len(segments) {
		return node.endpoint
	}

	seg := segments[depth]

	// 1. Try exact match first.
	if child, ok := node.children[seg]; ok {
		if result := pm.matchNode(child, segments, depth+1); result != nil {
			return result
		}
	}

	// 2. Try wildcard match.
	if node.wildcard != nil {
		if result := pm.matchNode(node.wildcard, segments, depth+1); result != nil {
			return result
		}
	}

	// 3. Try catch-all.
	if node.catchAll && node.endpoint != nil {
		return node.endpoint
	}

	return nil
}

// splitPath splits a URI path into segments, stripping leading/trailing slashes.
func splitPath(path string) []string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
