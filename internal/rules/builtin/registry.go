// Package builtin 包含全部内置规则，每条规则一个文件，样本在 testdata/<规则编号>/（MOD-05）。
package builtin

import "harness/internal/rules/engine"

// Rules 返回全部内置规则；每条规则一行，供 app 在启动时注册（IF-014）。
func Rules() []engine.Rule {
	return []engine.Rule{
		prd01{},
		prd02{},
		prd03{},
		prd04{},
		prd05{},
		prd06{},
		prd07{},
		prd08{},
		prd09{},
		prd10{},
		prd11{},
		term01{},
		term02{},
		arch01{},
		arch02{},
		arch03{},
		arch04{},
		arch05{},
		trace01{},
		trace02{},
		trace03{},
		adr01{},
		adr02{},
		adr03{},
		test01{},
		test02{},
		debt01{},
		debt02{},
		debt03{},
		task01{},
		task02{},
		task03{},
		task04{},
		task05{},
		task06{},
		commit01{},
		commit02{},
		commit03{},
		commit04{},
		approve01{},
	}
}
