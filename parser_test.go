package uyaml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParserBasic(t *testing.T) {
	output, err := parsePath("projects.(project='foo').version")
	require.NoError(t, err)
	require.Len(t, output, 3)
	require.Equal(t, output[0], pathKey("projects"))
	require.Equal(t, output[1], pathSelector{"project", "foo"})
	require.Equal(t, output[2], pathKey("version"))
}

func TestParserSimple(t *testing.T) {
	output, err := parsePath("projects")
	require.NoError(t, err)
	require.Len(t, output, 1)
	require.Equal(t, output[0], pathKey("projects"))
}

func TestParserSelector(t *testing.T) {
	output, err := parsePath("projects.(project='foo')")
	require.NoError(t, err)
	require.Len(t, output, 2)
	require.Equal(t, output[0], pathKey("projects"))
	require.Equal(t, output[1], pathSelector{"project", "foo"})
}

func TestParserInvalid(t *testing.T) {
	_, err := parsePath("projects(project='foo')")
	require.Error(t, err)
}
