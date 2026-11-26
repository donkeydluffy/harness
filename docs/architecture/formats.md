# 产物格式规范

- Status: Draft
- Updated: 2026-11-26
- Relates: ADR-002, MOD-03

所有产物都是 Markdown，由 MOD-03 的同一个解析器读取。机器只读取本文件定义的结构，其余文字自由书写。

## 1. 实体

实体是带编号的标题，加上紧随其后的字段列表：

```markdown
### MOD-04 engine
- Status: Active
- Depends: MOD-09
- Covers: FR-028, FR-045

自由文字……
```

- 标题级别不限（`##` 到 `####`），标题文字以编号开头，编号后是名称。
- 字段：标题之后、第一个空行之前，形如 `- Key: value` 的行。键为英文，区分大小写。
- 列表值用英文逗号分隔；空值表示空列表。
- 实体范围：从标题到下一个同级或更高级标题。

## 2. 编号

| 前缀 | 含义 | 正则 | 所在文件 |
|---|---|---|---|
| FR / NFR | 需求 | `FR-\d{3,}` / `NFR-\d{3,}` | `docs/PRD.md`、`docs/prd/*.md` |
| AC | 验收标准 | `AC-N?\d{3,}-\d+` | 需求内的列表项 `- AC-...:` |
| 用例 | 验收用例 | `<AC编号>/[NBE]\d+` | 验收标准下缩进两格的列表项；文字以 `Removed` 开头表示已删除 |
| TERM | 术语 | `TERM-\d{3,}` | PRD 统一语言表格的第一列 |
| OQ | 待澄清项 | `OQ-\d{3,}` | PRD |
| CHG | 变更记录 | `CHG-\d{3,}` | PRD 变更日志表格的第一列；同一编号可占多行 |
| MOD / EXT | 模块 / 扩展点 | `MOD-\d{2,}` / `EXT-\d{2,}` | `docs/architecture/overview.md` |
| IF | 接口 | `IF-\d{3,}` | `docs/architecture/modules/*.md` |
| FLOW | 调用流程 | `FLOW-\d{2,}` | `docs/architecture/flows.md` |
| ADR | 架构决策 | `ADR-\d{3,}` | `docs/adr/ADR-*.md`，一文件一条 |
| T | 任务 | `T-\d{3,}` | `docs/tasks.md` |
| DEBT | 技术债 | `DEBT-\d{3,}` | `docs/debt.md` |
| 规则 | 检查规则 | `[a-z]+\.\d{2,}`；项目规则同一形式，前缀不得是内置规则的区域（如 `acme.01`，Release 1 为 `x\.[a-z]+\.\d{2,}`） | 代码与 `harness.json` |

规则编号使用小写加点（如 `trace.01`），与产物编号（如 `ADR-001`）在形式上不重叠。

## 3. 各产物的必填字段

| 实体 | 必填字段 | 取值 |
|---|---|---|
| FR / NFR | Status, Release, Priority | Active / Removed；正整数；P0 / P1 / P2 |
| NFR | Type | Quantitative / Constraint / Scenario |
| FR / NFR（可选） | Replaces, Replaced-By | Replaces 列出被取代的需求或验收标准编号；被取代的需求用 Replaced-By 指回取代者，两边须互相对应（prd.12）。取代者的 Release 不大于 Current-Release 时，被取代的条目按 Removed 判定：覆盖类规则（trace.01、task.02、test.02）不再要求它，trace.03 报告对它的引用；在此之前被取代的条目照常有效 |
| NFR（Scenario） | Adds, Allowed-Paths, Closed-Modules | 文字 |
| OQ | Status；Closed 时 Resolution；可选 Relates（关联的需求编号，task.07 据此判定） | Open / Closed |
| MOD | Status, Paths, Public, Depends, Interfaces, Covers, Core, Coverage | Paths/Public 为 glob 列表；Core 为 true/false；Coverage 为 0-100 |
| IF | Module, Covers，及一个首行为 `// IF-xxx` 的代码块（签名） | |
| MOD / IF（可选） | Domain | 该模块或接口使用的领域概念代码名，逗号分隔；每个名称必须出现在统一语言表格的"代码名"列（规则 term.02） |
| FLOW | Covers, Steps | Steps 为 `MOD-a -> IF-xxx` 的有序列表，见第 4 节 |
| EXT | Status, Interface, Paths, Closed, Relates | Relates 必须是 Scenario 类 NFR |
| ADR | Status, Relates；正文含 `## 背景`、`## 决策`、`## 后果` | Proposed / Accepted / Superseded；Superseded 时 Superseded-By |
| T | Status, Covers, Design, Scope, Depends；可选 Extends, Flow | Status: Todo / InProgress / Done / Blocked；Covers: FR、NFR 或 AC 编号；Scope: 仓库相对路径 glob |
| T（机械改动，可选） | Kind, Rewrite, Relates | Kind: mechanical；Rewrite: 逗号分隔的 `<旧前缀> -> <新前缀>`；Relates 须含一个 ADR；此时 Design 可为空（AC-050-1） |
| CHG（接口变更） | 说明列以 `Breaking:` 开头 | 删除接口或修改签名的变更必须标注（AC-052-4） |
| DEBT | Status, Module | Open / Resolved |
| 进度记录 | 标题 `### <日期> T-n`（日期在前，不构成任务实体；标题后可加说明），正文含 `- 结论：` 与 `- 遗留：` 两行 | 位于 `docs/progress.md`；同一任务可有多条，任意一条两行齐全即可 |
| 产物文件 | 文件头 Status | Draft / Approved（Approved 只能由 `harness approve` 写入） |

## 4. 调用流程步骤

```markdown
### FLOW-02 写操作门禁
- Covers: FR-012, FR-025
- Steps: MOD-12 -> IF-017, MOD-11 -> IF-016, MOD-11 -> IF-012
```

每一步读作"调用方模块 -> 被调用接口"。规则 `arch.03` 检查：接口存在，且接口所属模块是调用方自身或在调用方的 Depends 中。

## 5. 测试引用编号

测试名中出现的用例、接口、扩展点编号视为引用。编号中的 `-` 和 `/` 在函数名中写作 `_`：

| 引用 | Go 测试名示例 |
|---|---|
| `AC-001-1/N1` | `TestInit_AC_001_1_N1` |
| `IF-012` | `TestRun_IF_012` |
| `EXT-01` | `TestExample_EXT_01` |

## 6. 项目配置 `harness.json`

```json
{
  "languages": ["go"],
  "sourcePaths": ["cmd/**", "internal/**"],
  "testReports": [".harness/reports/*.xml", ".harness/reports/*.jsonl"],
  "smoke": "go build ./... && go test ./internal/app/...",
  "sensitivePaths": [".env*", "*.pem", "*.key", "id_rsa*"],
  "sensitiveAllow": [".env.example"],
  "dangerousCommands": ["git push --force", "git push -f", "curl * | sh", "wget * | sh"],
  "vagueWords": ["快速", "友好", "适当", "尽量", "等等", "之类", "如有必要", "和/或", "大约", "灵活", "高效", "包括但不限于"],
  "commitMaxLines": 400,
  "stopMaxBlocks": 3,
  "staleCommits": 10,
  "start": "go run ./cmd/harness version",
  "tools": { "go": { "lint": "staticcheck ./..." } },
  "rules": []
}
```

整个文件的规范化哈希记入锁文件（规则 `cfg.01`）：按 JSON 解析，去掉键为 `"//"` 的注释项，键排序后紧凑序列化再求哈希，因此空白、键顺序和注释的变化不算配置变化。配置有实际变化时，须在 PRD 变更日志新增一行对象为 `harness.json` 的 CHG，或新增一条 Relates 含 `harness.json` 的 ADR，然后运行 `harness lock`。

`ci.01`、`cfg.02` 只检查存在 `harness.json` 的项目（执行过 `harness init`）。

Release 2 新增与删除的字段：

- `start`：启动命令；没有可启动应用的项目写 `none: <原因>`（AC-055-1）。`smoke` 为冒烟测试命令。
- `tools`：按语言覆盖 `format`、`lint` 命令，未覆盖的使用语言配置包的默认值；语言必须在 `languages` 中（AC-014-3，规则 cfg.03）。
- `rules`：声明式自定义规则，字段见 IF-036 的 `Def`（如 `{"id": "acme.01", "kind": "forbid", "paths": ["internal/**"], "pattern": "fmt\\.Println", "message": "用日志", "samples": ".harness/rules/acme.01", "approveFor": ["design"]}`）。
- `phases` 删除；存在时被忽略。

以下 `phases` 格式只在 Release 1 有效，Release 2 随 MOD-10 删除。

`phases` 中每项是以下两种之一（规则 `phase.01` 校验）：

- 新增阶段：`{"name": "release", "after": "implementation", "entry": [], "exit": ["prd.*"], "writable": ["docs/release/**"], "skill": "release"}`。`name` 不得与已有阶段同名；`entry`、`exit`、`writable`、`skill` 必须出现，`entry` 可以为空列表。
- 修改已有阶段：`{"name": "architecture", "addExit": ["x.lint"], "removeExit": ["adr.*"], "adr": "ADR-009"}`。只允许 `addEntry`、`addExit`、`removeEntry`、`removeExit`；有移除项时 `adr` 必须引用存在的 ADR。

规则引用写规则编号、`区域.*`（该区域全部规则）或 `*`（全部规则）。内置阶段文件 `internal/phase/builtin/<阶段>.json` 使用与"新增阶段"相同的字段（不含 `after`），顺序由 `order.json` 决定。当前阶段记录在 `harness.lock` 的 `phase` 字段，为空时是第一个阶段。

## 7. 锁文件 `harness.lock`

JSON，键按字典序排序，两次生成逐字节一致。

```json
{
  "version": 2,
  "items": { "FR-001": "sha256:…", "AC-001-1/N1": "sha256:…", "IF-001": "sha256:…", "snapshot:IF-001": "sha256:…", "rule:acme.01": "sha256:…", "config": "sha256:…" },
  "units": {
    "FR-001": { "hash": "sha256:…", "text": "### FR-001 …", "by": "name <email>", "at": "2026-11-25T10:00:00Z", "upstream": {} },
    "MOD-02": { "hash": "sha256:…", "text": "…", "by": "…", "at": "…", "upstream": { "FR-001": "sha256:…" } }
  },
  "audit": {}
}
```

规范化：统一为 LF，去掉行尾空白，连续空行合并为一行，去掉首尾空行。实体的哈希覆盖其标题到实体范围结束的全部文本。

`items` 的键与值：

| 键 | 值 |
|---|---|
| FR / NFR 编号 | 需求实体的哈希；需求为 Removed 时值前加 `removed:` |
| 用例编号 | 用例所在行文字的哈希；用例已删除时值前加 `removed:` |
| IF 编号 | 接口实体的哈希 |
| CHG 编号 | 该编号全部表格行的哈希，用于判断哪些变更记录是锁定之后新增的 |
| ADR 编号 | ADR 实体的哈希，用于判断哪些 ADR 是锁定之后新增的 |
| `config` | `harness.json` 的规范化哈希（文件存在时，见第 6 节） |
| `snapshot:IF-n` | 接口快照中该接口签名的哈希（AC-052-4） |
| `rule:<编号>` | 自定义规则的样本目录（全部文件按路径排序）的哈希（AC-054-2） |

`units` 按批准单位编号记录批准（AC-047-1）：`hash` 与 `text` 为批准时的规范化内容（不含 Status 字段；PRD 正文不含变更日志），`upstream` 为批准时各上游单位的哈希。Release 1 的 `phase` 与 `approvals` 由 `harness lock` 迁移后删除（AC-047-6、AC-048-1）。

锁文件不存在时视为空基线：依赖锁文件的规则不报告"消失"或"未登记"，`harness lock` 生成第一份锁文件。

## 8. 接口锚点与快照（Release 2，Go）

接口实现处的声明紧前一行写锚点注释：

```go
// harness:IF-022
func Check(root string, opt CheckOptions) (engine.Report, error) {
```

`harness snapshot` 把每个锚点后的声明去掉注释与函数体、空白规范化后写入 `docs/architecture/interfaces.txt`，每个接口一段、按编号排序：

```
IF-022
func Check(root string, opt CheckOptions) (engine.Report, error)
```

详细设计中接口代码块里同名声明经相同规范化后与快照比对（规则 if.03）；代码块中的说明注释不参与比对。

## 9. 外部工具报告

`harness check lint`、`harness check unit` 分别写入 `.harness/reports/lint.json`、`.harness/reports/unit.json`（字段见 IF-033），其中 `sourceHash` 为生成时 sourcePaths 内源文件的哈希，与当前不同即过期。报告不提交到仓库；CI 中先运行这两个命令再运行 `harness check`。
