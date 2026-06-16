package home

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/soulteary/flare/internal/fn"
	"github.com/soulteary/flare/internal/i18n"
)

// envCookieMaxAge is how long the chosen environment override persists, in seconds (30 days).
const envCookieMaxAge = 30 * 24 * 3600

// envCookieName is the cookie key carrying the user's network-environment override.
const envCookieName = "flare_env"

// RegisterEnvRouting exposes /env for setting the flare_env cookie that toggles bookmark URL resolution.
// GET /env?set=lan|wan — sets the cookie then 302-redirects to Referer (or "/" if none).
// Unknown / missing values redirect to "/" without touching the cookie.
func RegisterEnvRouting(e *echo.Echo) {
	e.GET("/env", envHandler)
}

func envHandler(c *echo.Context) error {
	set := strings.ToLower(strings.TrimSpace(c.QueryParam("set")))

	var cookieValue string
	var maxAge int
	switch set {
	case "lan":
		cookieValue = "lan"
		maxAge = envCookieMaxAge
	case "wan":
		cookieValue = "wan"
		maxAge = envCookieMaxAge
	default:
		return c.Redirect(http.StatusFound, "/")
	}

	c.SetCookie(&http.Cookie{
		Name:     envCookieName,
		Value:    cookieValue,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})

	referer := c.Request().Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	return c.Redirect(http.StatusFound, referer)
}

// readEnvMode returns the user's preferred environment from the flare_env cookie.
// Falls back to fn.EnvLAN when the cookie is missing or carries an unknown value.
func readEnvMode(r *http.Request) fn.EnvMode {
	if r == nil {
		return fn.EnvLAN
	}
	c, err := r.Cookie(envCookieName)
	if err != nil {
		return fn.EnvLAN
	}
	mode := fn.ParseEnvMode(c.Value)
	if mode == fn.EnvAuto {
		// The toggle UI only exposes lan/wan; legacy "auto" or garbage values default to LAN.
		return fn.EnvLAN
	}
	return mode
}

// putEnvToggleURL sets m["EnvToggleURL"] to the /env target that flips the current mode
// (e.g. "/env?set=wan" when currently lan, and vice versa). Call before rendering any page
// that includes the toolbar.
func putEnvToggleURL(m map[string]interface{}, current fn.EnvMode) {
	if current == fn.EnvWAN {
		m["EnvToggleURL"] = "/env?set=lan"
	} else {
		m["EnvToggleURL"] = "/env?set=wan"
	}
}

// putEnvToggleFields fills EnvToggleURL and EnvToggleTitle for the toolbar toggle button.
// EnvToggleURL is the /env target that flips the current mode; EnvToggleTitle is the hover hint
// naming the *target* mode (e.g. "切换到公网" when currently on LAN). Call before rendering any
// page that includes the toolbar.
func putEnvToggleFields(m map[string]interface{}, current fn.EnvMode, locale string) {
	if current == fn.EnvWAN {
		m["EnvToggleURL"] = "/env?set=lan"
		m["EnvToggleTitle"] = i18n.T(locale, "env_toggle_to_lan")
	} else {
		m["EnvToggleURL"] = "/env?set=wan"
		m["EnvToggleTitle"] = i18n.T(locale, "env_toggle_to_wan")
	}
}
