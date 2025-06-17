#!/usr/bin/env python3
import argparse, json, os, signal, sys, time
from camoufox import launch


def parse():
    p = argparse.ArgumentParser()
    p.add_argument("--url", required=True, help="page to keep in browser")
    p.add_argument("--cookie", default="./data/cookie.json")
    p.add_argument("--bootstrap", action="store_true",
                   help="open GUI, wait manual login, save cookie & exit")
    p.add_argument("--proxy", help="http(s) proxy, e.g. http://user:pass@ip:port")
    p.add_argument("--headless", action="store_true", default=False)
    return p.parse_args()


def main():
    a = parse()
    opts = {
        "headless": a.headless,
        "proxy": {"server": a.proxy} if a.proxy else None,
        "virtual_display": not a.headless,
    }
    browser = launch(**opts)
    ctx = browser.new_context()

    if os.path.exists(a.cookie):
        with open(a.cookie) as f:
            ctx.add_cookies(json.load(f))
    page = ctx.new_page()
    page.goto(a.url)

    if a.bootstrap:
        print(">> \u624b\u52a8\u5b8c\u6210 Google \u767b\u5f55\u540e\u6309 Ctrl-C \u7ed3\u675f")
        try:
            while True:
                time.sleep(1)
        finally:
            with open(a.cookie, "w") as f:
                json.dump(ctx.cookies(), f)
            print(f">> cookie \u4fdd\u5b58\u5230 {a.cookie}")
    else:
        signal.pause()


if __name__ == "__main__":
    main()
