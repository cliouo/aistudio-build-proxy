package browser

import (
	"log"

	"github.com/playwright-community/playwright-go"
)

type Driver struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

func New(ws string) (*Driver, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	b, err := pw.Firefox.Connect(ws)
	if err != nil {
		_ = pw.Stop()
		return nil, err
	}

	return &Driver{pw: pw, browser: b}, nil
}

func (d *Driver) KeepAlive(url, storage string) error {
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

	if _, err = page.Goto(url); err != nil {
		return err
	}

	_, err = page.WaitForEvent("websocket", playwright.PageWaitForEventOptions{
		Timeout: playwright.Float(30 * 1000),
	})
	if err != nil {
		return err
	}

	log.Printf("Camoufox page connected, url=%s", url)

	_, err = page.WaitForEvent("close")
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) Close() {
	_ = d.browser.Close()
	_ = d.pw.Stop()
}
