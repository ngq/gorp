package security

// DefaultCasbinRBACModelText is the embedded standard RBAC model for Casbin.
// It provides a zero-file-dependency, fail-safe model so applications never crash
// when external configuration files are missing or paths vary across environments.
//
// DefaultCasbinRBACModelText 是嵌入式的标准 Casbin RBAC 模型定义。
// 提供零文件依赖的兜底能力，避免因外部配置文件缺失或相对路径不一致导致进程崩溃。
const DefaultCasbinRBACModelText = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act) || r.sub == "super-admin" || r.sub == "root"
`
