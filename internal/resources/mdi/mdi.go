package mdi

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/faninx/flare/config/define"
)

// _CACHE_MDI_ICON_DATA caches inline SVG snippets keyed by icon name.
var _CACHE_MDI_ICON_DATA = make(map[string]string)

//go:embed mdi-cheat-sheets
var MdiExampleAssets embed.FS

func RegisterRouting(e *echo.Echo) {
	if mdiExample, err := fs.Sub(MdiExampleAssets, "mdi-cheat-sheets"); err == nil {
		e.StaticFS(define.RegularPages.Icons.Path, mdiExample)
	}
}

const _EMPTY_ICON = ""

// GetIconByName returns an inline <svg> snippet for the named Material Design Icon.
// Previously this function also wrote each icon into a memfs and rendered an <img>
// pointing to a sub-URL; that path was dropped because memfs.File did not implement
// io.ReadSeeker and Echo v5's StaticFS returned 500 for every icon request.
func GetIconByName(name string) string {
	if name == "" {
		return _EMPTY_ICON
	}
	if cached, ok := _CACHE_MDI_ICON_DATA[name]; ok {
		return cached
	}
	icon := iconMap[strings.ToLower(name)]
	if icon == "" {
		_CACHE_MDI_ICON_DATA[name] = _EMPTY_ICON
		return _EMPTY_ICON
	}
	content := `<svg viewBox="0 0 24 24"><path d="` + icon + `" style="fill: var(--color-primary);"></path></svg>`
	_CACHE_MDI_ICON_DATA[name] = content
	return content
}
