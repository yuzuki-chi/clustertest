package api

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestPveClient_buildUrl(t *testing.T) {
	t.Run("normal_case", func(t *testing.T) {
		c := &PveClient{
			PveClientOption: PveClientOption{
				Address: "base",
			},
		}
		assert.Equal(t, "base/path", c.buildUrl("path"))
	})
	t.Run("slash", func(t *testing.T) {
		c := &PveClient{
			PveClientOption: PveClientOption{
				Address: "base/",
			},
		}
		assert.Equal(t, "base/path", c.buildUrl("/path"))
	})
}
