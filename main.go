package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"
	//	"encoding/hex"
	//	"io"
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/go-ini/ini"
	"github.com/jroimartin/gocui"
)

var (
	connection  net.Conn
	cmdLine     *gocui.View
	usrList     *gocui.View
	channelView *gocui.View
	theme       Theme
	conf        Config
	handle      string
	beepfunc    *syscall.Proc
	host        string
	port        int
	cmdBuffer   []string
)

// main ...
func main() {
	msgBoxInit()
	initFlash()

	//MessageBox("Are you?", "Are you cool?", MB_YESNO)
	beepfunc = syscall.MustLoadDLL("user32.dll").MustFindProc("MessageBeep")
	//beepfunc.Call(0xffffffff)

	flag.StringVar(&handle, "handle", "", "Defines your handle or 'nickname'.")
	flag.StringVar(&host, "host", "", "JRC Server IP/Hostname.")
	flag.IntVar(&port, "port", -1, "JRC Server port.")

	flag.Parse()

	conf = DefaultConfig()
	theme = DefaultTheme()

	cfg, _ := ini.Load("settings.ini")
	cfg.Section("server").HasKey("host")

	loadConfig(&conf, &theme)

	servchan := make(chan []byte)

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.Mouse = true

	g.SetManagerFunc(layout)

	// Ctrl-C quick quit.
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// Bind the enter key to submit to the server.
	if err = g.SetKeybinding("cmdl", gocui.KeyEnter, gocui.ModNone, submit); err != nil {
		log.Panicln("Can't bind the enter key.")
	}

	go handleServerMessages(g, servchan)

	// Connect to server.
	connection, err = net.Dial("tcp", fmt.Sprintf("%v:%v", conf.host, conf.port))
	if err != nil {
		os.Exit(1)
	}

	//handle = "Jeff"
	helo()

	go func() {

		cbuf := bufio.NewReader(connection)
		for {
			line, _, err := cbuf.ReadLine()
			if err != nil {
				break
			}
			servchan <- line
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

// loadConfig ...
func loadConfig(c *Config, t *Theme) {
	cfg, err := ini.Load("settings.ini")
	if err != nil {
		os.Exit(1)
	}

	// host
	if host != "" {
		c.host = host
	} else if cfg.Section("server").HasKey("host") {
		c.host = cfg.Section("server").Key("host").String()
	}

	// port
	if port != -1 {
		c.port = port
	} else if cfg.Section("server").HasKey("port") {
		c.port, _ = cfg.Section("server").Key("port").Int()
	}

	// handle
	if handle != "" {
		c.handle = handle
	} else if cfg.Section("client").HasKey("handle") {
		c.handle = cfg.Section("client").Key("handle").String()
	}

	if cfg.Section("client").HasKey("message_sound") {
		c.msgSound, _ = cfg.Section("client").Key("message_sound").Bool()
	}

	LoadTheme(cfg, t)
}

// layout ...
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Channel window.
	channelView, err := g.SetView("chan", 0, 0, maxX-17, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	channelView.Autoscroll = true
	channelView.Editable = false

	channelView.Wrap = true
	channelView.Title = "#channel"

	// User list
	usrList, err = g.SetView("usrl", maxX-17, 0, maxX-1, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	// Command line bar.
	cmdLine, err = g.SetView("cmdl", 0, maxY-3, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	cmdLine.Editable = true

	//

	_, err = g.SetCurrentView("cmdl")
	if err != nil {
		return err
	}

	return nil
}

// quit ...
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// helo ...
func helo() error {
	n, err := connection.Write([]byte(fmt.Sprintf("ident %s\n", conf.handle)))
	if err != nil || n == 0 {
		return err
	}
	return nil
}

// submit ...
func submit(g *gocui.Gui, v *gocui.View) error {

	cmdLine.Rewind()

	var cmd string
	rawCmd := strings.TrimSpace(cmdLine.Buffer())
	cmd = rawCmd

	// Prepare the command for the server
	if strings.HasPrefix(rawCmd, "/") {
		cmd = parseCmd(g, rawCmd)
	} else {
		cmd = fmt.Sprintf("MSG %v", rawCmd)
	}

	if len(cmd) < 1 {
		cmdLine.Clear()
		resetCursor()
		return nil
	}

	// Write to jrc server
	n, err := connection.Write([]byte(cmd + "\n"))
	if err != nil || n == 0 {
		return err
	}

	// wipe buffer
	cmdLine.Clear()

	// reset cursor
	err = cmdLine.SetCursor(0, 0)
	if err != nil {
		return err
	}

	return nil
}

func resetCursor() {
	// reset cursor
	_ = cmdLine.SetCursor(0, 0)
}

func parseCmd(g *gocui.Gui, rawCmd string) string {
	ch, _ := g.View("chan")
	cmdParm := strings.Split(rawCmd, " ")
	switch cmdParm[0] {
	case "/quit":
		os.Exit(0)

	case "/nick":
		if len(cmdParm) > 1 {
			return fmt.Sprintf("NICK %v", cmdParm[1])
		}

	case "/away":
		if len(cmdParm) > 1 {
			return fmt.Sprintf("AWAY %v", strings.TrimPrefix(rawCmd, "/away "))
		}
		return "AWAY"

	case "/me":
		if len(cmdParm) > 1 {
			return fmt.Sprintf("ACTION %v", strings.TrimPrefix(rawCmd, "/me "))
		}

	case "/sendurl":
		if len(cmdParm) > 2 {
			return fmt.Sprintf("URL %v %v", cmdParm[1] /* target */, cmdParm[2] /* url */)
		}

	case "/clear":
		ch.Clear()

	case "/savebuffer":
		t := time.Now()
		filename := fmt.Sprintf("%v-capturedBuffer.txt", t.Format("2006-01-02_15-04-05"))

		if len(cmdParm) > 1 {
			filename = cmdParm[1]
		}

		err := ioutil.WriteFile(filename, []byte(ch.Buffer()), 0644)
		if err != nil {
			fmt.Fprintf(ch, "- %v\n", err.Error())
		}

		fmt.Fprintf(ch, "- Saved buffer as %v\n", filename)
	}

	return ""
}

// handleServerMessages ...
func handleServerMessages(g *gocui.Gui, servchan <-chan []byte) {
	for {
		select {
		case line := <-servchan:
			v, _ := g.View("chan")

			if strings.HasPrefix(string(line), "open.url") {

				pParms := strings.Split(string(line), "|")
				if len(pParms) != 3 {
					//fmt.Fprintf(v, "%+v\n", pParms)
				}

				from := pParms[1]
				url := pParms[2]

				result := MessageBox("PUSH", fmt.Sprintf("%v wants to send you to:\n%v\nDo it?", from, url), MB_YESNO)
				line = []byte(fmt.Sprintf("- Launching url: %v", url))
				if result == 6 {
					err := exec.Command("rundll32", "url.dll,FileProtocolHandler", strings.TrimSpace(url)).Start()
					if err != nil {
						line = []byte(fmt.Sprintf("! Failed to launch url: %v", err.Error()))
					}
				} else {
					line = []byte(fmt.Sprintf("- Launch url cancelled."))
				}
			}

			if strings.HasPrefix(string(line), "sound.alert") {
				parms := strings.Split(string(line), "|")
				if len(parms) != 2 {
					//fmt.Fprintf(v, "%+v\n")
				}
				from := parms[1]

				line = []byte(fmt.Sprintf("- %v is trying to get your attention!", from))
				beepfunc.Call(0xffffffff)
			}

			fmt.Fprint(v, colorizeLine(line))

			if conf.msgSound {
				if !strings.HasPrefix(string(line), conf.handle) {
					beepfunc.Call(0xffffffff)

				}
			}

			if strings.HasPrefix(string(line), "*") {
				flashWindow(true)
			}

			g.Execute(func(g *gocui.Gui) error { return nil }) // Hack to force UI update
		}
	}
}

// colorizeLine
func colorizeLine(line []byte) string {

	ln := strings.TrimSpace(string(line))
	// server notice
	if strings.HasPrefix(string(line), "- ") {
		return fmt.Sprintf(
			"\x1b[3%d;%dm%s\x1b[0m\n",
			theme.srvNotice,
			theme.srvNoticeStyle,
			line,
		)
	}

	// join / part
	if strings.HasPrefix(string(line), "*** ") {
		return fmt.Sprintf(
			"\x1b[3%d;%dm%s\x1b[0m\n",
			theme.chnJoinPart,
			theme.chnJoinPartStyle,
			line,
		)
	}

	// away / back
	if strings.HasPrefix(string(line), "** ") {
		return fmt.Sprintf(
			"\x1b[3%d;%dm%s\x1b[0m\n",
			theme.chnJoinPart,
			theme.chnJoinPartStyle,
			line,
		)
	}

	// action eg: /me
	if strings.HasPrefix(string(line), "* ") {
		return fmt.Sprintf(
			"\x1b[3%d;%dm%s\x1b[0m\n",
			theme.chnAction,
			theme.chnActionStyle,
			line,
		)
	}

	// errors
	if strings.HasPrefix(string(line), "! ") {
		return fmt.Sprintf(
			"\x1b[3%d;%dm%s\x1b[0m\n",
			theme.chatError,
			theme.chatErrorStyle,
			line,
		)
	}

	// messages
	return fmt.Sprintf(
		"\x1b[3%d;%dm%s\x1b[0m\n",
		theme.chnMessage,
		theme.chnMessageStyle,
		ln,
	)
}
