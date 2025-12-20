package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

type ScrapeResult struct {
	URL        string
	HTML       string
	Screenshot []byte
	Links      []string
	StatusCode int
	Error      error
	ScrapedAt  time.Time
}

func main() {
	outputDir := flag.String("output", "output", "Cikti dosyalarinin kaydedilecegi dizin")
	timeout := flag.Int("timeout", 60, "Islem zaman asimi (saniye)")
	flag.Parse()

	args := flag.Args()
	var targetURL string
	if len(args) > 0 {
		targetURL = args[0]
	}

	if targetURL == "" {
		fmt.Println("GO WEB SCRAPER - CTI Araci")
		fmt.Println("===========================")
		fmt.Println("")
		fmt.Println("Kullanim:")
		fmt.Println("  ./web-scraper <url>")
		fmt.Println("  ./web-scraper <url> -output <dizin>")
		fmt.Println("")
		fmt.Println("Parametreler:")
		fmt.Println("  <url>     : Taranacak web sitesi adresi (zorunlu)")
		fmt.Println("  -output   : Cikti dizini (varsayilan: output)")
		fmt.Println("  -timeout  : Zaman asimi saniye (varsayilan: 60)")
		fmt.Println("")
		fmt.Println("Ornek:")
		fmt.Println("  ./web-scraper https://example.com")
		fmt.Println("  ./web-scraper https://example.com -output ./data")
		os.Exit(1)
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		log.Fatalf("[HATA] Gecersiz URL formati. HTTP veya HTTPS ile baslamalidir: %s", targetURL)
	}

	finalOutputDir := getOutputPath(*outputDir, parsedURL)

	fmt.Println("GO WEB SCRAPER - CTI Araci")
	fmt.Println("===========================")
	fmt.Printf("Hedef URL: %s\n", targetURL)
	fmt.Printf("Cikti Dizini: %s\n", finalOutputDir)
	fmt.Printf("Zaman Asimi: %d saniye\n\n", *timeout)

	if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
		log.Fatalf("[HATA] Cikti dizini olusturulamadi: %v", err)
	}

	result := scrapeWebsite(targetURL, *timeout)

	if result.Error != nil {
		log.Fatalf("[HATA] Scraping hatasi: %v", result.Error)
	}

	saveResults(result, finalOutputDir)

	fmt.Println("\nIslem basariyla tamamlandi.")
}

func getOutputPath(baseDir string, parsedURL *url.URL) string {
	host := parsedURL.Hostname()

	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")
	var domain, subdomain string

	if len(parts) >= 2 {
		domain = parts[len(parts)-2] + "-" + parts[len(parts)-1]

		if len(parts) > 2 {
			subParts := parts[:len(parts)-2]
			subdomain = strings.Join(subParts, "-")
		}
	} else {
		domain = host
	}

	if subdomain != "" {
		return filepath.Join(baseDir, domain, subdomain)
	}
	return filepath.Join(baseDir, domain)
}

func getVersionedFilename(dir, filename string) string {
	fullPath := filepath.Join(dir, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fullPath
	}

	version := 2
	for {
		versionedName := fmt.Sprintf("%d-%s", version, filename)
		versionedPath := filepath.Join(dir, versionedName)

		if _, err := os.Stat(versionedPath); os.IsNotExist(err) {
			return versionedPath
		}
		version++
	}
}

func scrapeWebsite(targetURL string, timeoutSeconds int) ScrapeResult {
	result := ScrapeResult{
		URL:       targetURL,
		ScrapedAt: time.Now(),
	}

	fmt.Println("[*] Tarayici baslatiliyor...")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx,
		chromedp.WithErrorf(func(s string, i ...interface{}) {}),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	var htmlContent string
	var screenshot []byte

	fmt.Println("[*] Sayfaya baglaniliyor...")

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		chromedp.FullScreenshot(&screenshot, 90),
	)

	if err != nil {
		result.Error = fmt.Errorf("sayfa yuklenirken hata olustu: %w", err)
		return result
	}

	fmt.Println("[+] Sayfa basariyla yuklendi")
	fmt.Println("[+] Ekran goruntusu alindi")

	result.HTML = htmlContent
	result.Screenshot = screenshot

	fmt.Println("[*] Sayfadaki linkler cikariliyor...")
	result.Links = extractLinks(htmlContent, targetURL)
	fmt.Printf("[+] %d adet link bulundu\n", len(result.Links))

	return result
}

func extractLinks(htmlContent string, baseURL string) []string {
	links := make([]string, 0)
	uniqueLinks := make(map[string]bool)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("[!] HTML parse hatasi: %v", err)
		return links
	}

	base, _ := url.Parse(baseURL)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || href == "#" {
			return
		}

		parsedHref, err := url.Parse(href)
		if err != nil {
			return
		}

		absoluteURL := base.ResolveReference(parsedHref).String()

		if !uniqueLinks[absoluteURL] {
			uniqueLinks[absoluteURL] = true
			links = append(links, absoluteURL)
		}
	})

	return links
}

func saveResults(result ScrapeResult, outputDir string) {
	fmt.Println("\n[*] Sonuclar kaydediliyor...")

	htmlPath := getVersionedFilename(outputDir, "page_content.html")
	if err := os.WriteFile(htmlPath, []byte(result.HTML), 0644); err != nil {
		log.Printf("[!] HTML dosyasi kaydedilemedi: %v", err)
	} else {
		fmt.Printf("[+] HTML icerigi kaydedildi: %s\n", htmlPath)
	}

	screenshotPath := getVersionedFilename(outputDir, "screenshot.png")
	if err := os.WriteFile(screenshotPath, result.Screenshot, 0644); err != nil {
		log.Printf("[!] Ekran goruntusu kaydedilemedi: %v", err)
	} else {
		fmt.Printf("[+] Ekran goruntusu kaydedildi: %s\n", screenshotPath)
	}

	linksPath := getVersionedFilename(outputDir, "links.txt")
	linksContent := strings.Join(result.Links, "\n")
	if err := os.WriteFile(linksPath, []byte(linksContent), 0644); err != nil {
		log.Printf("[!] Link listesi kaydedilemedi: %v", err)
	} else {
		fmt.Printf("[+] Link listesi kaydedildi: %s\n", linksPath)
	}
}
