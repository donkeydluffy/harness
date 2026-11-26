# 架构总览

- Status: Draft

### MOD-01 api
- Status: Active
- Paths: internal/api/**
- Public: internal/api/api.go
- Depends: MOD-02
- Interfaces: IF-001
- Covers: FR-001
- Core: true
- Coverage: 80

### MOD-02 store
- Status: Active
- Paths: internal/store/**
- Public: internal/store/store.go
- Depends:
- Interfaces: IF-002
- Covers: FR-001, FR-099
- Core: false
- Coverage: 70

### EXT-01 新增存储
- Status: Active
- Interface: IF-002
- Paths: internal/store/<实现>/**
- Closed: MOD-01
- Relates: NFR-001
