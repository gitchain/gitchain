package git

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

const fixtureCommit = `tree 69218c749588a7147b99ff45bf7d18db1bb126e8
parent d3030ad8f6ad49e5ad69a2842f06940c60f9db6f
author Yurii Rashkovskii <yrashk@gmail.com> 1400767572 +0800
committer Yurii Rashkovskii <yrashk@gmail.com> 1400767572 +0800

Add HACKING.md`

func TestCommitDecode(t *testing.T) {
	c := &Commit{}
	c.SetBytes([]byte(fixtureCommit))
	assert.Equal(t, hex.EncodeToString(c.Tree), "69218c749588a7147b99ff45bf7d18db1bb126e8")
	assert.Equal(t, hex.EncodeToString(c.Parent), "d3030ad8f6ad49e5ad69a2842f06940c60f9db6f")
	assert.Equal(t, c.Author, "Yurii Rashkovskii <yrashk@gmail.com> 1400767572 +0800")
	assert.Equal(t, c.Committer, "Yurii Rashkovskii <yrashk@gmail.com> 1400767572 +0800")
	assert.Equal(t, c.Message, "Add HACKING.md")
}
