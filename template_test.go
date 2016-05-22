package main

import (
	"testing"

	"github.com/commondream/yamlast"
	"github.com/stvp/assert"
)

func TestMetadata(t *testing.T) {
	config := LoadConfig()
	template := NewTemplate(config)
	template.LoadFile("fixtures/template/metadata.yml")

	node := yamlast.SelectNode(template.DocumentNode,
		"Resources.SomeResource.Metadata")
	assert.NotNil(t, node)
	assert.Equal(t, yamlast.ScalarNode, node.Kind)
	assert.Equal(t, "Rad", node.Value)
}
