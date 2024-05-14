package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	info := extractAppInfo("svc.state.demoService.default.10.10.10.10:7654")
	assert.True(t, info != nil)
	const nsStr2 = `"app.ns.appId.groupId.default.props.10.232.123.98:23"`
	info2 := extractNamespaceInfo(nsStr2)
	assert.True(t, info2 != nil)
}
