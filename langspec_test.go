package teal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLangSpec(t *testing.T) {
	assert.Equal(t, BuiltInLangSpecs[len(BuiltInLangSpecs)-1].EvalMaxVersion, LatestLangSpec.EvalMaxVersion)
}
