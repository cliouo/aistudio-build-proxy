package browser

import (
	"fmt"
	"log"
	"os"

	"github.com/playwright-community/playwright-go"
)

type Driver struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

type KeepAliveOption func(*playwright.BrowserNewContextOptions, playwright.Page)

// WithAuthToken injects an auth token into localStorage.
func WithAuthToken(token string) KeepAliveOption {
	return func(opts *playwright.BrowserNewContextOptions, page playwright.Page) {
		if token == "" {
			return
		}
		script := playwright.Script{Content: playwright.String(fmt.Sprintf(`localStorage.setItem('auth_token', '%s');`, token))}
		if err := page.Context().AddInitScript(script); err == nil {
			log.Println("Injected auth_token into localStorage.")
		}
	}
}

// New connects to an existing Playwright browser via WebSocket.
func New(ws string) (*Driver, error) {
	if ws == "" {
		endpointFile := "/tmp/endpoint.txt"
		if os.Getenv("GOOS") == "windows" {
			endpointFile = os.Getenv("TEMP") + "\\endpoint.txt"
		}
		b, err := os.ReadFile(endpointFile)
		if err != nil {
			return nil, fmt.Errorf("CAMOUFOX_WS not set and failed to read endpoint: %w", err)
		}
		ws = string(b)
	}
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	b, err := pw.Firefox.Connect(ws)
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("could not connect to browser: %w", err)
	}
	return &Driver{pw: pw, browser: b}, nil
}

// KeepAlive loads the given URL and waits until the page is closed.
func (d *Driver) KeepAlive(url, storage string, opts ...KeepAliveOption) error {
	ctxOpts := playwright.BrowserNewContextOptions{
		StorageStatePath: playwright.String(storage),
	}
	ctx, err := d.browser.NewContext(ctxOpts)
	if err != nil {
		return fmt.Errorf("could not create context: %w", err)
	}
	defer ctx.Close()
	page, err := ctx.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}
	for _, o := range opts {
		o(&ctxOpts, page)
	}
	log.Printf("Navigating to URL: %s", url)
	if _, err = page.Goto(url); err != nil {
		return fmt.Errorf("could not navigate: %w", err)
	}
	_, err = page.WaitForEvent("close")
	return err
}

func (d *Driver) SaveStorageState(path string) error {
	if len(d.browser.Contexts()) == 0 {
		return fmt.Errorf("no browser contexts open")
	}
	ctx := d.browser.Contexts()[0]
	_, err := ctx.StorageState(path)
	return err
}

func (d *Driver) Close() {
	if d.browser != nil {
		d.browser.Close()
	}
	if d.pw != nil {
		d.pw.Stop()
	}
}
