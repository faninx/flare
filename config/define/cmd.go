package define

import (
	"github.com/faninx/flare/config/model"
)

const (
	DEFAULT_PORT                     = 5005
	DEFAULT_ENABLE_GUIDE             = true
	DEFAULT_ENABLE_DEPRECATED_NOTICE = true
	DEFAULT_DISABLE_LOGIN            = true
	DEFAULT_ENABLE_OFFLINE           = false
	DEFAULT_USER_NAME                = "flare"
	DEFAULT_ENABLE_EDITOR            = true
	DEFAULT_VISIBILITY               = "DEFAULT"
	DEFAULT_DISABLE_CSP              = false

	DEFAULT_COOKIE_NAME   = "flare"
	DEFAULT_COOKIE_SECRET = "secret"
	// DEFAULT_COOKIE_SECURE 默认 false：flare 的常见部署是「内网 HTTP / 公网经反代 HTTPS」，
	// 把 Secure 设成 true 会让浏览器在 LAN HTTP 下不回传 session cookie。需要严格 HTTPS-only 时
	// 显式把 FLARE_COOKIE_SECURE 设为 true 即可。
	DEFAULT_COOKIE_SECURE = false
)

// get default env config
func GetDefaultEnvVars() model.Envs {
	return model.Envs{
		Port:                   DEFAULT_PORT,
		EnableGuide:            DEFAULT_ENABLE_GUIDE,
		EnableDeprecatedNotice: DEFAULT_ENABLE_DEPRECATED_NOTICE,
		DisableLoginMode:       DEFAULT_DISABLE_LOGIN,
		EnableOfflineMode:      DEFAULT_ENABLE_OFFLINE,
		EnableEditor:           DEFAULT_ENABLE_EDITOR,
		Visibility:             DEFAULT_VISIBILITY,
		DisableCSP:             DEFAULT_DISABLE_CSP,

		User: DEFAULT_USER_NAME,
		Pass: "",

		CookieName:   DEFAULT_COOKIE_NAME,
		CookieSecret: DEFAULT_COOKIE_SECRET,
		CookieSecure: DEFAULT_COOKIE_SECURE,
	}
}

var DefaultEnvVars = GetDefaultEnvVars()

var AppFlags model.Flags

// FLARE_VISIBLE defines visibility levels: "DEFAULT" or "PRIVATE".
var FLARE_VISIBLE = "PRIVATE"
