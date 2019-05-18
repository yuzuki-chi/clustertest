package config

import "testing"

func TestLoadFromBytes(t *testing.T) {
	t.Run("empty_config", func(t *testing.T) {
		conf, err := LoadFromBytes([]byte(""))
		if err != nil {
			t.Errorf("LoadFromBytes() returns an error: %s", err.Error())
			t.Fail()
			return
		}
		t.Logf("conf: %s", conf.String())
	})
}
