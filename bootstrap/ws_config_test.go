package bootstrap

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestWSConfig_OptionalAndExplicitBools(t *testing.T) {
	v := viper.New()
	v.SetConfigType("env")
	if err := v.ReadConfig(strings.NewReader("HTTP_PORT=8080\n")); err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	var env Environment
	if err := v.Unmarshal(&env); err != nil {
		t.Fatalf("Unmarshal empty WS failed: %v", err)
	}
	if env.WS.Enabled != nil || env.WS.AuthRequired != nil {
		t.Fatalf("expected nil pointers when WS_* keys are unset, got %+v", env.WS)
	}

	v2 := viper.New()
	v2.SetConfigType("env")
	cfgText := "WS_ENABLED=false\nWS_AUTH_REQUIRED=false\nWS_WRITE_TIMEOUT=7s\n"
	if err := v2.ReadConfig(strings.NewReader(cfgText)); err != nil {
		t.Fatalf("ReadConfig 2 failed: %v", err)
	}

	var env2 Environment
	if err := v2.Unmarshal(&env2); err != nil {
		t.Fatalf("Unmarshal explicit WS failed: %v", err)
	}
	if env2.WS.Enabled == nil || *env2.WS.Enabled != false {
		t.Fatalf("expected WS_ENABLED=false, got %v", env2.WS.Enabled)
	}
	if env2.WS.AuthRequired == nil || *env2.WS.AuthRequired != false {
		t.Fatalf("expected WS_AUTH_REQUIRED=false, got %v", env2.WS.AuthRequired)
	}
	if env2.WS.WriteTimeout != 7*time.Second {
		t.Fatalf("expected WS_WRITE_TIMEOUT=7s, got %v", env2.WS.WriteTimeout)
	}
}
