package gate

import (
	"math/rand"
	"testing"
)

// 由英文单词拼成的长标识符不判为密钥；随机 token 几乎都仍判为密钥（T-036）。
func TestWordIdentifierNotSecret(t *testing.T) {
	// 拼接构造，以免测试文件本身被门禁拦下
	for _, s := range []string{
		"TestParseGoJSON" + "PackageFailure" + "_IF_008",
		"BenchmarkHTTP2Server" + "StreamingResponse_1KB",
		"TestParseJUnit" + "NestedSuites" + "WithSkipped_IF_008",
	} {
		if id, _ := findSecret("func " + s + "(t *testing.T) {}"); id != "" {
			t.Errorf("%s 不应判为密钥（%s）", s, id)
		}
	}
	for _, s := range []string{
		"ghp_" + "aB3xY9kLmN2p" + "Q7rS4tU6vW8x" + "Z1cD5eF0gH",
		"sk_live_" + "4eC39HqLyjWDarjtT1" + "zdp7dc9Zk2Lm",
		"wJalrXUtnFEMI" + "/K7MDENG/" + "bPxRfiCYEXAMPLEKEY",
	} {
		if id, _ := findSecret("token = " + s); id != "secret.entropy" {
			t.Errorf("%s 应判为高熵 token，实际 %q", s, id)
		}
	}
	const cs = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"
	r := rand.New(rand.NewSource(1))
	total, missed := 0, 0
	for i := 0; i < 200000; i++ {
		b := make([]byte, 32+r.Intn(40))
		for j := range b {
			b[j] = cs[r.Intn(len(cs))]
		}
		if tok := string(b); highEntropy(tok) {
			total++
			if wordIdentifier(tok) {
				missed++
			}
		}
	}
	if missed*20000 > total { // 上限万分之 0.5
		t.Fatalf("%d/%d 个随机 token 被当作标识符放过", missed, total)
	}
}
