// Package lock 负责内容规范化与哈希、harness.lock 读写、批准记录（MOD-06）。
package lock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
)

// File 是锁文件名，位于项目根目录。
const File = "harness.lock"

// RemovedPrefix 标记已 Removed 的需求或用例的哈希值。
const RemovedPrefix = "removed:"

// Normalize 统一换行为 LF，去掉行尾空白，合并连续空行，去掉首尾空行（IF-005）。
//
// harness:IF-005
func Normalize(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	var out []string
	blank := false
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimRight(l, " \t")
		if l == "" {
			if blank || len(out) == 0 {
				continue
			}
			blank = true
		} else {
			blank = false
		}
		out = append(out, l)
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, "\n")
}

// Hash 返回 "sha256:<hex>"，输入先经 Normalize（IF-005）。
func Hash(text string) string {
	sum := sha256.Sum256([]byte(Normalize(text)))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Lock 是 harness.lock 的内容（IF-006）。Version 为 2 时按批准单位记录批准（Units，ADR-009）；
// 读到第 1 版时 Legacy 是 Release 1 的按文件批准，只供 harness lock 迁移，迁移后为空、不再写出。
// Release 1 的 phase 字段读到时忽略，写出时不再出现（AC-048-1）。
type Lock struct {
	Version int                        `json:"version"`
	Items   map[string]string          `json:"items"`
	Units   map[string]UnitApproval    `json:"units,omitempty"`
	Legacy  []Approval                 `json:"approvals,omitempty"`
	Audit   map[string]json.RawMessage `json:"audit"`
}

// UnitApproval 是一个批准单位的批准记录（AC-047-1）：批准时的规范化内容与哈希、批准人、时间，
// 以及批准时各上游单位的哈希。
type UnitApproval struct {
	Hash     string            `json:"hash"`
	Text     string            `json:"text"`
	By       string            `json:"by"`
	At       time.Time         `json:"at"`
	Upstream map[string]string `json:"upstream"`
}

// UnitVersion 是按批准单位记录批准的锁文件版本。
const UnitVersion = 2

// ByUnit 判断锁文件是否按批准单位记录批准。
func (l *Lock) ByUnit() bool { return l.Version >= UnitVersion }

// Approval 是 Release 1 的一次按文件批准记录。
type Approval struct {
	Artifact string    `json:"artifact"`
	Hash     string    `json:"hash"`
	By       string    `json:"by"`
	At       time.Time `json:"at"`
}

// New 返回空锁文件（第 2 版）。
func New() *Lock {
	return &Lock{Version: UnitVersion, Items: map[string]string{}, Units: map[string]UnitApproval{}, Legacy: []Approval{}, Audit: map[string]json.RawMessage{}}
}

// Read 读取锁文件；文件不存在时返回空 Lock 而非错误（IF-006）。
//
// harness:IF-006
func Read(path string) (*Lock, error) {
	src, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		l := New()
		l.Version = 1 // 尚未锁定：没有批准记录可比对，按 Release 1 的规则只检查文件头
		return l, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取锁文件 %s：%w", path, err)
	}
	l := New()
	l.Version = 1 // 文件中没有 version 时是 Release 1 的格式
	if err := json.Unmarshal(src, l); err != nil {
		return nil, fmt.Errorf("解析锁文件 %s：%w", path, err)
	}
	if l.Items == nil {
		l.Items = map[string]string{}
	}
	if l.Legacy == nil {
		l.Legacy = []Approval{}
	}
	if l.Units == nil && l.ByUnit() {
		l.Units = map[string]UnitApproval{}
	}
	if l.Audit == nil {
		l.Audit = map[string]json.RawMessage{}
	}
	return l, nil
}

// Exists 判断锁文件是否存在。
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Marshal 输出锁文件内容：全部对象的键按字典序，两空格缩进，末尾换行。
func (l *Lock) Marshal() ([]byte, error) {
	raw, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	// 经通用结构再编码一次，使结构体字段也按键排序。
	var generic any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&generic); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(generic); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Write 写出锁文件；两次写出逐字节一致（IF-006）。
func (l *Lock) Write(path string) error {
	b, err := l.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// Approved 判断产物的当前内容是否有有效批准：最近一次批准的哈希与当前哈希一致（IF-006）。
func (l *Lock) Approved(artifact, currentHash string) bool {
	for i := len(l.Legacy) - 1; i >= 0; i-- {
		if l.Legacy[i].Artifact == artifact {
			return l.Legacy[i].Hash == currentHash
		}
	}
	return false
}

// LastApproval 返回产物最近一次批准记录，没有时返回 nil。
func (l *Lock) LastApproval(artifact string) *Approval {
	for i := len(l.Legacy) - 1; i >= 0; i-- {
		if l.Legacy[i].Artifact == artifact {
			return &l.Legacy[i]
		}
	}
	return nil
}
