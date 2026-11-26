package builtin

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd08：内容哈希相对锁文件变化的需求必须被锁定之后新增的 CHG 条目引用，且动作与差异一致。
type prd08 struct{}

func (prd08) ID() string       { return "prd.08" }
func (prd08) Covers() []string { return []string{"AC-004-1", "AC-004-2"} }
func (prd08) Fix() string {
	return "在 PRD 变更日志中新增一行 CHG（新编号），动作写新增、修改或删除，对象写变化的需求编号；确认后运行 harness lock"
}

var (
	reqIDRe    = regexp.MustCompile(`^(FR|NFR)-\d+$`)
	reqRangeRe = regexp.MustCompile(`^((?:FR|NFR)-)(\d+)\.\.((?:FR|NFR)-)(\d+)$`)
	anyIDRe    = regexp.MustCompile(`^[A-Z]+-N?\d+`)
)

const (
	diffAdd    = "新增"
	diffModify = "修改"
	diffRemove = "删除"
)

func (prd08) Check(s *workspace.Snapshot) []engine.Violation {
	base := s.Lock.Items
	if len(base) == 0 {
		return nil // 空基线：首次 harness lock 之前不追踪变更
	}
	cur := s.LockItems()
	diff := map[string]string{}
	for id, now := range cur {
		if !reqIDRe.MatchString(id) {
			continue
		}
		was, ok := base[id]
		switch {
		case !ok:
			diff[id] = diffAdd
		case was == now:
		case !strings.HasPrefix(was, lock.RemovedPrefix) && strings.HasPrefix(now, lock.RemovedPrefix):
			diff[id] = diffRemove
		default:
			diff[id] = diffModify
		}
	}
	var out []engine.Violation
	declared := map[string]bool{}
	for _, c := range s.Project.Changes {
		if _, old := base[c.ID]; old || !chgRe.MatchString(c.ID) {
			continue
		}
		if c.Action != diffAdd && c.Action != diffModify && c.Action != diffRemove {
			out = append(out, engine.Violation{File: c.File, Line: c.Line,
				Msg: fmt.Sprintf("%s 的动作 %q 不合法，合法取值：新增、修改、删除", c.ID, c.Action)})
			continue
		}
		for _, target := range c.Targets {
			if target == workspace.ConfigFile {
				continue // 门禁配置的变更由 cfg.01 检查
			}
			ids, ok := expandTarget(target)
			if !ok {
				out = append(out, engine.Violation{File: c.File, Line: c.Line,
					Msg: fmt.Sprintf("%s 的对象 %q 不是编号", c.ID, target)})
				continue
			}
			for _, id := range ids {
				if !reqIDRe.MatchString(id) {
					continue // 术语、接口等对象不在本规则追踪范围
				}
				declared[id] = true
				if msg := mismatch(id, c.Action, diff[id]); msg != "" {
					out = append(out, engine.Violation{File: c.File, Line: c.Line, Msg: c.ID + "：" + msg})
				}
			}
		}
	}
	var changed []string
	for id := range diff {
		if !declared[id] {
			changed = append(changed, id)
		}
	}
	sort.Strings(changed)
	for _, id := range changed {
		file, line := locate(s, id)
		out = append(out, engine.Violation{File: file, Line: line,
			Msg: fmt.Sprintf("%s 相对锁文件有%s，但变更日志中没有引用它的新 CHG 条目", id, diff[id])})
	}
	return out
}

var chgRe = regexp.MustCompile(`^CHG-\d{3,}$`)

func mismatch(id, declared, actual string) string {
	switch {
	case declared == actual:
		return ""
	case actual == "":
		return fmt.Sprintf("声明%s %s，但 %s 内容未变", declared, id, id)
	case actual == diffRemove:
		return fmt.Sprintf("声明%s %s，但 %s 实际被改为 Removed，应为删除", declared, id, id)
	case actual == diffAdd:
		return fmt.Sprintf("声明%s %s，但 %s 是锁文件中不存在的新需求，应为新增", declared, id, id)
	default:
		return fmt.Sprintf("声明%s %s，但 %s 实际是内容修改，应为修改", declared, id, id)
	}
}

// expandTarget 把对象展开为编号列表，支持 FR-001..FR-005 区间；不是编号时返回 false。
func expandTarget(t string) ([]string, bool) {
	if m := reqRangeRe.FindStringSubmatch(t); m != nil && m[1] == m[3] {
		a, _ := strconv.Atoi(m[2])
		b, _ := strconv.Atoi(m[4])
		var out []string
		for i := a; i <= b; i++ {
			out = append(out, fmt.Sprintf("%s%0*d", m[1], len(m[2]), i))
		}
		return out, true
	}
	if anyIDRe.MatchString(t) && !strings.ContainsAny(t, " \t") {
		return []string{t}, true
	}
	return nil, false
}
