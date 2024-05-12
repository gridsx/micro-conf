package cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testYaml = `
app:
  id: demoService
id: 123121
users:
  - name: Lucy
    age: 12
  - name: Tom
    age: 22
`

func TestYaml(t *testing.T) {
	m, err := YamlToKvList(testYaml)
	assert.True(t, err == nil)
	assert.True(t, len(m) > 0)

	_, err = YamlToProperties(testYaml)
	assert.True(t, err == nil)

	j, err := YamlToJson(testYaml)
	assert.True(t, err == nil)
	assert.True(t, IsJson(j))
}

func TestJson(t *testing.T) {
	j, err := YamlToJson(testYaml)
	assert.True(t, err == nil)
	y, err := JsonToYaml(j)
	assert.True(t, err == nil)
	m, err := YamlToKvList(y)
	assert.True(t, err == nil)
	assert.True(t, len(m) > 0)
}

func TestProps(t *testing.T) {
	p, err := YamlToProperties(testYaml)
	assert.True(t, err == nil)
	assert.True(t, len(p) > 0)
	m, err := PropertiesToMap(p)
	assert.True(t, err == nil)
	assert.True(t, len(m) > 0)
}
