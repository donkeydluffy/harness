package testreport

import (
	"strings"
	"testing"
)

func TestParseJUnit_IF_008(t *testing.T) {
	src := `<testsuites><testsuite name="s1">
<testcase name="a" classname="pkg.A" file="a_test.py"/>
<testcase name="b" classname="pkg.A"><failure message="x"/></testcase>
<testcase name="c"><skipped/></testcase>
<testsuite name="nested"><testcase name="d"><error/></testcase></testsuite>
</testsuite></testsuites>`
	got, err := ParseJUnit(strings.NewReader(src))
	if err != nil || len(got) != 4 {
		t.Fatalf("%+v %v", got, err)
	}
	want := []TestCase{{"a", "pkg.A", "a_test.py", Passed}, {"b", "pkg.A", "", Failed}, {"c", "s1", "", Skipped}, {"d", "nested", "", Failed}}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%d: %+v want %+v", i, got[i], want[i])
		}
	}
	if _, err := ParseJUnit(strings.NewReader("<html/>")); err == nil {
		t.Fatal("want root error")
	}
}

func TestParseGoJSON_IF_008(t *testing.T) {
	src := `{"Action":"run","Package":"m/a","Test":"TestA"}
{"Action":"output","Package":"m/a","Test":"TestA","Output":"x"}
{"Action":"pass","Package":"m/a","Test":"TestA/sub"}
{"Action":"fail","Package":"m/a","Test":"TestA"}
{"Action":"skip","Package":"m/b","Test":"TestB"}
# m/c build failed
{"Action":"pass","Package":"m/a"}
`
	got, err := ParseGoJSON(strings.NewReader(src))
	if err != nil || len(got) != 3 {
		t.Fatalf("%+v %v", got, err)
	}
	if got[0] != (TestCase{"TestA/sub", "m/a", "", Passed}) || got[1].Status != Failed || got[2].Status != Skipped {
		t.Fatalf("%+v", got)
	}
}

// 包失败而其中没有失败的测试时记一条包级失败；包因测试失败而失败时不重复记录（T-034）。
func TestGoJSONPkgFail_IF_008(t *testing.T) {
	src := `{"Action":"start","Package":"m/build"}
{"Action":"output","Package":"m/build","Output":"# m/build\n"}
{"Action":"fail","Package":"m/build","FailedBuild":"m/build"}
{"Action":"pass","Package":"m/main","Test":"TestOK"}
{"Action":"fail","Package":"m/main"}
{"Action":"fail","Package":"m/test","Test":"TestBad"}
{"Action":"fail","Package":"m/test"}
{"Action":"pass","Package":"m/ok","Test":"TestFine"}
{"Action":"pass","Package":"m/ok"}
`
	got, err := ParseGoJSON(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	var pkgFail []string
	for _, tc := range got {
		if tc.Name == PackageFailure {
			if tc.Status != Failed {
				t.Errorf("包级记录应为 failed：%+v", tc)
			}
			pkgFail = append(pkgFail, tc.Suite)
		}
	}
	// m/build 构建失败、m/main 的测试都通过但包失败（如 TestMain），各记一条；m/test 已有失败的测试，m/ok 通过。
	if strings.Join(pkgFail, ",") != "m/build,m/main" {
		t.Fatalf("包级失败：%v（全部结果 %+v）", pkgFail, got)
	}
}
