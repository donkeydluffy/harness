# ADR-016 Repo 提供最后一次改动路径的提交

- Status: Proposed
- Relates: FR-030, IF-007

## 背景

doc.05（AC-030-4）要统计模块代码在其详细设计最后一次修改之后的提交次数。IF-007 只有 `CommitsTouching(paths, since)`，`since` 须是提交，而 Repo 没有办法得到某个文件最后一次被改动的提交。

## 决策

IF-007 新增 `LastCommit(paths []string) (string, error)`，用 `git log -1 --format=%H -- <paths>` 实现；doc.05 先取详细设计的最后提交，再用 `CommitsTouching(模块 Paths, 该提交)` 计数。

被否决的方案：只用 `CommitsTouching` 对提交历史二分查找最后一次改动，每个模块要数十次 git 调用；或由 workspace 直接调用 git，绕开 vcs 模块。

## 后果

IF-007 的实现方（vcs）与测试替身须提供该方法；下游 MOD-05、MOD-09 只新增调用，已有调用不变。
