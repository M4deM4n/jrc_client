package main

import (
	"github.com/go-ini/ini"
)

const COLOR_RED int = 1
const COLOR_GREEN int = 2
const COLOR_YELLOW int = 3
const COLOR_BLUE int = 4
const COLOR_PINK int = 5
const COLOR_TEAL int = 6
const COLOR_WHITE int = 7

const STYLE_BRIGHT int = 1
const STYLE_NORMAL int = 5
const STYLE_BACKGR int = 7

// Theme contains color references for various message types.
type Theme struct {
	chnAction        int
	chnActionStyle   int
	chnJoinPart      int
	chnJoinPartStyle int
	chnMessage       int
	chnMessageStyle  int
	chatError        int
	chatErrorStyle   int
	srvNotice        int
	srvNoticeStyle   int
}

// DefaultTheme returns a Theme with preset values.
func DefaultTheme() Theme {
	t := Theme{}
	t.srvNotice = COLOR_TEAL
	t.srvNoticeStyle = STYLE_NORMAL
	t.chnJoinPart = COLOR_WHITE
	t.chnJoinPartStyle = STYLE_BRIGHT
	t.chnAction = COLOR_PINK
	t.chnActionStyle = STYLE_BRIGHT
	t.chnMessage = COLOR_WHITE
	t.chnMessageStyle = STYLE_NORMAL
	t.chatError = COLOR_RED
	t.chatErrorStyle = STYLE_BRIGHT

	return t
}

// ...
func LoadTheme(cfg *ini.File, t *Theme) {
	if cfg.Section("theme").HasKey("error") {
		t.chatError, _ = cfg.Section("theme").Key("error").Int()
	}

	if cfg.Section("theme").HasKey("error_style") {
		t.chatError, _ = cfg.Section("theme").Key("error_style").Int()
	}

	if cfg.Section("theme").HasKey("channel_action") {
		t.chnAction, _ = cfg.Section("theme").Key("channel_action").Int()
	}

	if cfg.Section("theme").HasKey("channel_action_style") {
		t.chnActionStyle, _ = cfg.Section("theme").Key("channel_action_style").Int()
	}

	if cfg.Section("theme").HasKey("channel_join_part") {
		t.chnJoinPart, _ = cfg.Section("theme").Key("channel_join_part").Int()
	}

	if cfg.Section("theme").HasKey("channel_join_part_style") {
		t.chnJoinPartStyle, _ = cfg.Section("theme").Key("channel_join_part_style").Int()
	}

	if cfg.Section("theme").HasKey("channel_message") {
		t.chnMessage, _ = cfg.Section("theme").Key("channel_message").Int()
	}

	if cfg.Section("theme").HasKey("channel_message_style") {
		t.chnMessageStyle, _ = cfg.Section("theme").Key("channel_message_style").Int()
	}

	if cfg.Section("theme").HasKey("server_notice") {
		t.srvNotice, _ = cfg.Section("theme").Key("server_notice").Int()
	}

	if cfg.Section("theme").HasKey("server_notice_style") {
		t.srvNoticeStyle, _ = cfg.Section("theme").Key("server_notice_style").Int()
	}
}
