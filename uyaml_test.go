package uyaml

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const yamlFile = `usersCount: 2
users:
  - name: josie
    roles:
      - bot
      - foo
      - bar
    admin: true
    createdAt: 0
    weight: 1.3
  - name: lester
    roles:
      - dummy
`

func TestDecode(t *testing.T) {
	type userStruct struct {
		Name string `yaml:"name"`
		Roles []string `yaml:"roles"`
		Admin bool `yaml:"admin"`
		CreatedAt int `yaml:"createdAt"`
		Weight float64 `yaml:"weight"`
	}
	d, err := Decode([]byte(yamlFile))
	require.NoError(t, err)
	require.NotNil(t, d)
	ok, v, err := d.DigItem("users.(name='josie')")
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.NoError(t, err)
	var usr userStruct
	err = v.Decode(&usr)
	require.NoError(t, err)
	assert.Equal(t, userStruct{
		Name:      "josie",
		Roles:     []string{"bot", "foo", "bar"},
		Admin:     true,
		CreatedAt: 0,
		Weight:    1.3,
	}, usr)
	ok, m := v.Map()
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{
		"name":      "josie",
		"roles":     []interface{}{"bot", "foo", "bar"},
		"admin":     true,
		"createdAt": int64(0),
		"weight":    1.3,
	}, m)
}

func TestRemove(t *testing.T) {
	d, err := Decode([]byte(yamlFile))
	require.NoError(t, err)
	require.NotNil(t, d)
	ok, v, err := d.DigItem("users.(name='josie')")
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.NoError(t, err)
	err = v.Remove()
	assert.NoError(t, err)
	b, err := d.Encode()
	assert.NoError(t, err)
	assert.Equal(t, "usersCount: 2\nusers:\n  - name: lester\n    roles:\n      - dummy\n", string(b))
}

func TestReplace(t *testing.T) {
	d, err := Decode([]byte(yamlFile))
	require.NoError(t, err)
	require.NotNil(t, d)
	ok, v, err := d.DigItem("users.(name='josie')")
	assert.True(t, ok)
	assert.NotNil(t, v)
	assert.NoError(t, err)
	_, err = v.Replace(map[string]interface{}{
		"name": "n",
		"test": true,
	})
	assert.NoError(t, err)
	b, err := d.Encode()
	assert.NoError(t, err)
	assert.Equal(t, "usersCount: 2\nusers:\n  - name: n\n    test: true\n  - name: lester\n    roles:\n      - dummy\n", string(b))
}

func TestCreateStructure(t *testing.T) {
	d, err := Decode([]byte(yamlFile))
	require.NoError(t, err)
	require.NotNil(t, d)
	v, err := d.Set("admins.(name='josie')", map[string]interface{}{"test": true})
	assert.NotNil(t, v)
	assert.NoError(t, err)
	b, err := d.Encode()
	assert.NoError(t, err)
	assert.Equal(t, "usersCount: 2\nusers:\n  - name: josie\n    roles:\n      - bot\n      - foo\n      - bar\n    admin: true\n    createdAt: 0\n    weight: 1.3\n  - name: lester\n    roles:\n      - dummy\nadmins:\n  - name: josie\n    test: true\n", string(b))
}
