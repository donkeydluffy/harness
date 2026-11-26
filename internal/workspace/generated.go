package workspace

// RegeneratedFiles 是 harness init 能按模板确定性重新生成、并由 cfg.02 校验内容的生成文件。
// 代理可以删除它们，再单独运行一次 harness init 写回（DEBT-009）。
// 不在其中的：hook 配置（.claude/settings.json）只能由人修改（AC-025-4），代理删除后不写回就会让下次会话失去门禁；
// harness.json 与 harness.lock 删除后重新生成会丢失项目配置与批准记录。
var RegeneratedFiles = []string{".github/workflows/harness.yml", ".gitlab-ci.yml"}
