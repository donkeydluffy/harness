package adapter

import (
	"testing"

	"harness/internal/gate"
)

type fakeAgent struct{}

func (fakeAgent) Name() string { return "fake" }
func (fakeAgent) Decode(event string, stdin []byte) (gate.Request, error) {
	return gate.Request{Kind: gate.Write, Paths: []string{string(stdin)}}, nil
}
func (fakeAgent) Encode(event string, d gate.Decision) ([]byte, int) {
	if d.Allow {
		return nil, 0
	}
	return []byte(d.Reason), 2
}
func (fakeAgent) HookConfig(binary string) (string, []byte) {
	return ".fake/hooks", []byte(binary + " hook fake")
}

// 演示 EXT-02：新增一种代理只需实现 Adapter 并在 registry.go 登记一行。
func TestExample_EXT_02(t *testing.T) {
	Register(fakeAgent{})
	a, ok := Get("fake")
	if !ok {
		t.Fatal("未注册")
	}
	req, _ := a.Decode("pre", []byte("src/a.go"))
	if req.Paths[0] != "src/a.go" {
		t.Fatalf("%+v", req)
	}
	if out, code := a.Encode("pre", gate.Decision{Reason: "no"}); code != 2 || string(out) != "no" {
		t.Fatal("encode")
	}
	if _, ok := Get("claudecode"); !ok || len(Names()) < 2 {
		t.Fatalf("内置适配器：%v", Names())
	}
}
