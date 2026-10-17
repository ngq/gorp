// Package security provides the minimal authentication and security contracts for gorp framework.
package security

// CaptchaStoreKey is the container key for the captcha store capability.
const CaptchaStoreKey = "framework.security.captcha_store"

// CaptchaStore defines the unified interface for captcha storage, compatible with base64Captcha.Store.
//
// CaptchaStore 定义统一的验证码存储接口，与 base64Captcha.Store 规范完全兼容。
type CaptchaStore interface {
	// Set stores the captcha value associated with the given id.
	// Set 存储与给定 id 关联的验证码值。
	Set(id string, value string) error

	// Get retrieves the captcha value associated with the given id.
	// If clear is true, the captcha value will be deleted after reading.
	// Get 获取与给定 id 关联的验证码值。如果 clear 为 true，则读取后清除。
	Get(id string, clear bool) string

	// Verify checks whether the answer matches the stored captcha value.
	// If clear is true, the captcha value will be deleted upon verification.
	// Verify 校验答案是否与存储的验证码相符。如果 clear 为 true，校验后清除。
	Verify(id, answer string, clear bool) bool
}
