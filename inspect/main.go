package main

import (
	"fmt"
	"log"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	if err := playwright.Install(); err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	if _, err = page.Goto("http://localhost/dashboard"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	// not authenticated yet -> redirected to login; sign in with demo creds
	if err := page.Locator("[data-testid=email-input]").Fill("demo@fintech.com"); err != nil {
		log.Fatalf("could not fill email: %v", err)
	}
	if err := page.Locator("[data-testid=password-input]").Fill("password123"); err != nil {
		log.Fatalf("could not fill password: %v", err)
	}
	if err := page.Locator("[data-testid=login-btn]").Click(); err != nil {
		log.Fatalf("could not click login: %v", err)
	}
	if err := page.WaitForURL("**/dashboard"); err != nil {
		log.Fatalf("did not land on dashboard: %v", err)
	}

	title, _ := page.Title()
	fmt.Println("=== TITLE ===")
	fmt.Println(title)

	fmt.Println("\n=== HEADINGS (h1-h3) ===")
	headings, _ := page.Locator("h1, h2, h3").AllTextContents()
	for _, h := range headings {
		fmt.Println("-", h)
	}

	fmt.Println("\n=== BUTTONS ===")
	buttons, _ := page.Locator("button").AllTextContents()
	for _, b := range buttons {
		fmt.Println("-", b)
	}

	fmt.Println("\n=== LINKS ===")
	links, _ := page.Locator("a").All()
	for _, l := range links {
		text, _ := l.TextContent()
		href, _ := l.GetAttribute("href")
		fmt.Printf("- %q -> %s\n", text, href)
	}

	fmt.Println("\n=== data-testid ELEMENTS ===")
	testids, _ := page.Locator("[data-testid]").All()
	for _, el := range testids {
		id, _ := el.GetAttribute("data-testid")
		text, _ := el.TextContent()
		fmt.Printf("- %s: %q\n", id, text)
	}
}
