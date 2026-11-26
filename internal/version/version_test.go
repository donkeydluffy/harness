package version

import (
	"runtime/debug"
	"testing"
)

func build(mod string, settings ...string) *debug.BuildInfo {
	info := &debug.BuildInfo{Main: debug.Module{Version: mod}}
	for i := 0; i+1 < len(settings); i += 2 {
		info.Settings = append(info.Settings, debug.BuildSetting{Key: settings[i], Value: settings[i+1]})
	}
	return info
}

func TestCurrent_IF_037(t *testing.T) {
	const rev = "d368ca65724147e8064ca93a7d3f20e275719047"
	cases := []struct {
		name, injected string
		info           *debug.BuildInfo
		want           string
	}{
		{"注入的发布版本号", "v0.1.4", build("(devel)"), "0.1.4"},
		{"模块版本", "", build("v0.2.0"), "0.2.0"},
		{"伪版本带未提交改动", "", build("v0.1.4-0.20261126111344-d368ca657241+dirty"), "0.1.4-0.20261126111344-d368ca657241-dirty"},
		{"没有模块版本时用提交哈希", "", build("(devel)", "vcs.revision", rev, "vcs.modified", "false"), "d368ca657241"},
		{"提交哈希带未提交改动", "", build("", "vcs.revision", rev, "vcs.modified", "true"), "d368ca657241-dirty"},
		{"都取不到", "", build("(devel)"), "dev"},
		{"没有构建信息", "", nil, "dev"},
	}
	for _, c := range cases {
		if got := fromBuild(c.injected, c.info); got != c.want {
			t.Errorf("%s：得到 %q，期望 %q", c.name, got, c.want)
		}
	}
	if Current() == "" {
		t.Error("Current 不应返回空串")
	}
}

func TestRelease_IF_037(t *testing.T) {
	for v, want := range map[string]bool{
		"0.2.0": true, "10.0.12": true,
		"dev": false, "d368ca657241": false, "0.2.0-dirty": false,
		"0.1.4-0.20261126111344-d368ca657241": false, "v0.2.0": false, "": false,
	} {
		if got := Release(v); got != want {
			t.Errorf("Release(%q) = %v，期望 %v", v, got, want)
		}
	}
}
