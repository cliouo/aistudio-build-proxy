package browser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// Driver wraps a playwright browser connection.
type Driver struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

// New connects to the Playwright server via WebSocket endpoint.
func ResolveEndpoint(ws string) string {
	if ws != "" {
		return ws
	}
	if data, err := os.ReadFile("/tmp/endpoint.txt"); err == nil {
		if ep := strings.TrimSpace(string(data)); ep != "" {
			return ep
		}
	}
	return "ws://localhost:38835"
}

func New(ws string) (*Driver, error) {
	ws = ResolveEndpoint(ws)
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	b, err := pw.Firefox.Connect(ws)
	if err != nil {
		pw.Stop()
		return nil, err
	}
	return &Driver{pw: pw, browser: b}, nil
}

// KeepAlive keeps the page alive using the provided storage state.
func (d *Driver) KeepAlive(url, storage string, bootstrap bool) error {
	ctx, err := d.browser.NewContext(playwright.BrowserNewContextOptions{
		StorageStatePath: playwright.String(storage),
	})
	if err != nil {
		return err
	}

	page, err := ctx.NewPage()
	if err != nil {
		return err
	}

	if token := os.Getenv("AUTH_TOKEN"); token != "" && !strings.Contains(url, "auth_token=") {
		page.AddInitScript(playwright.Script{Content: playwright.String(fmt.Sprintf(`localStorage.setItem("auth_token", "%s");`, token))})
	}

	if _, err = page.Goto(url); err != nil {
		return err
	}

	if bootstrap {
		log.Println("Bootstrap: waiting for manual login, press <Enter> when done…")
		go func() {
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			ctx.StorageState(storage)
			log.Printf("storage saved to %s, exiting…", storage)
			os.Exit(0)
		}()
	}

	_, err = page.WaitForEvent("websocket", playwright.PageWaitForEventOptions{
		Timeout: playwright.Float(30 * 1000),
	})
	if err != nil {
		return err
	}

	log.Printf("Camoufox page connected, url=%s", url)
	page.WaitForEvent("close")
	return io.EOF
}

// Close shuts down the browser connection.
func (d *Driver) Close() {
	d.browser.Close()
	d.pw.Stop()
}
