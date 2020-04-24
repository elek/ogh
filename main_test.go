package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseReference(t *testing.T) {
	str := "elek/repo@trunk#123"
	ref := ParseReference(str)
	assert.Equal(t, "elek", ref.Org)
	assert.Equal(t, "repo", ref.Repo)
	assert.Equal(t, "trunk", ref.Branch)
	assert.Equal(t, "123", ref.Id)

	str = "elek/repo@trunk"
	ref = ParseReference(str)
	assert.Equal(t, "elek", ref.Org)
	assert.Equal(t, "repo", ref.Repo)
	assert.Equal(t, "trunk", ref.Branch)
	assert.Equal(t, "", ref.Id)

	str = "elek/repo#123"
	ref = ParseReference(str)
	assert.Equal(t, "elek", ref.Org)
	assert.Equal(t, "repo", ref.Repo)
	assert.Equal(t, "master", ref.Branch)
	assert.Equal(t, "123", ref.Id)

	str = "elek/repo"
	ref = ParseReference(str)
	assert.Equal(t, "elek", ref.Org)
	assert.Equal(t, "repo", ref.Repo)
	assert.Equal(t, "master", ref.Branch)
	assert.Equal(t, "", ref.Id)

	str = "repo"
	ref = ParseReference(str)
	assert.Equal(t, "apache", ref.Org)
	assert.Equal(t, "repo", ref.Repo)
	assert.Equal(t, "master", ref.Branch)
	assert.Equal(t, "", ref.Id)
}
