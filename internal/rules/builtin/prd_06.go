package builtin

import (
	"fmt"
	"regexp"
	"strings"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// prd06：澄清记录按轮次标题（日期 + 第 N 轮）组织，每轮不超过 5 个问题。
type prd06 struct{}

func (prd06) ID() string       { return "prd.06" }
func (prd06) Covers() []string { return []string{"AC-002-1"} }
func (prd06) Fix() string {
	return "把问答放在 \"### YYYY-MM-DD 第 N 轮\" 标题下，每轮最多 5 条 \"- Q: … A: …\"；超出的问题移到下一轮"
}

var (
	roundDateRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	roundNumRe  = regexp.MustCompile(`第\s*\d+\s*轮`)
)

const maxQuestions = 5

func (prd06) Check(s *workspace.Snapshot) []engine.Violation {
	idx := s.Doc(workspace.PRDIndex)
	if idx == nil {
		return nil
	}
	start, end, ok := chapter(idx, "澄清记录")
	if !ok {
		return nil
	}
	var out []engine.Violation
	round, roundLine, count := "", 0, 0
	flush := func() {
		if count > maxQuestions {
			out = append(out, engine.Violation{File: idx.Path, Line: roundLine,
				Msg: fmt.Sprintf("轮次 %q 有 %d 个问题，超过 %d 个", round, count, maxQuestions)})
		}
	}
	for ln := start + 1; ln < end && ln <= len(idx.Lines); ln++ {
		line := idx.Lines[ln-1]
		if strings.HasPrefix(line, "### ") {
			flush()
			round, roundLine, count = strings.TrimPrefix(line, "### "), ln, 0
			if !roundDateRe.MatchString(round) || !roundNumRe.MatchString(round) {
				out = append(out, engine.Violation{File: idx.Path, Line: ln,
					Msg: fmt.Sprintf("轮次标题 %q 缺少日期或轮次（应为 YYYY-MM-DD 第 N 轮）", round)})
			}
			continue
		}
		if strings.HasPrefix(line, "- Q:") || strings.HasPrefix(line, "- Q：") {
			if roundLine == 0 {
				out = append(out, engine.Violation{File: idx.Path, Line: ln, Msg: "问答不在任何轮次标题下"})
			}
			count++
		}
	}
	flush()
	return out
}
