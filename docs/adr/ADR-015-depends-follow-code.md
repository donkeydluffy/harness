# ADR-015 模块 Depends 按代码的实际 import 修正

- Status: Proposed
- Relates: FR-053, MOD-01, MOD-02, MOD-11, MOD-13, MOD-17

## 背景

T-050 的规则 deps.01、deps.03 第一次对照本仓库的 Go import 与架构声明，发现 10 处不一致：cli 直接 import engine、phase、gate、scaffold；app 与 gate import model；scaffold import engine；而 app 对 lang、scaffold 对 phase、custom 对 model 的声明依赖在非测试代码中没有使用。

## 决策

- 按代码修正声明：MOD-01 加入 MOD-04、MOD-10、MOD-11、MOD-13；MOD-02 加入 MOD-03、去掉 MOD-14；MOD-11 加入 MOD-03；MOD-13 加入 MOD-04、去掉 MOD-10；MOD-17 去掉 MOD-03。依赖仍然无环。
- FLOW-07 去掉 `MOD-13 -> IF-015`；FLOW-12 中调用 IF-035 的是 MOD-09（app 经 workspace 提取锚点）。
- 被否决的方案：改代码使 cli 只经 app 调用下层模块。要改 cli、gate、scaffold 的多处调用，超出 T-050 的范围，收益只是依赖图更窄。

## 后果

- cli 的依赖面变宽：engine、gate、scaffold 的公开接口变化会直接影响 cli。
- 此后依赖的变化由 deps.01、deps.03 在每次全量检查中发现，修改 Depends 须同时登记 ADR（adr.03）。
