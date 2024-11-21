package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// HTTP client
var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
	},
	Timeout: 10 * time.Second,
}

func checkDependency(dep, url string, verbose bool) bool {
	apiURL := fmt.Sprintf("https://api.allorigins.win/raw?url=https://www.npmjs.com/search?q=%s", dep)
	resp, err := httpClient.Get(apiURL)
	if err != nil {
		if verbose {
			fmt.Printf("[ERROR] Failed to fetch dependency '%s' from URL '%s': %s\n", dep, url, err)
		}
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if verbose {
			fmt.Printf("[ERROR] Failed to read response for '%s' from URL '%s': %s\n", dep, url, err)
		}
		return false
	}

	// check the response body for the "0 packages found" string
	if strings.Contains(string(body), "0 packages found") {
		return false
	}
	return true
}

func processURL(url string, verbose bool, wg *sync.WaitGroup) {
	defer wg.Done()

	if verbose {
		fmt.Printf("[INFO] Fetching URL: %s\n", url)
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		if verbose {
			fmt.Printf("[ERROR] Failed to fetch '%s': %s\n", url, err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if verbose {
			fmt.Printf("[ERROR] Non-200 response for '%s': %d\n", url, resp.StatusCode)
		}
		return
	}

	var pkg PackageJSON
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		if verbose {
			fmt.Printf("[ERROR] Failed to parse JSON from '%s': %s\n", url, err)
		}
		return
	}

	// check dependencies and devDependencies concurrently
	var depWg sync.WaitGroup
	checkDep := func(dep string, dev bool) {
		defer depWg.Done()
		if !checkDependency(dep, url, verbose) {
			if dev {
				fmt.Printf("[WARNING] Potential Dependency Confusion: '%s' not found on npm (devDependency, %s)\n", dep, url)
			} else {
				fmt.Printf("[WARNING] Potential Dependency Confusion: '%s' not found on npm (%s)\n", dep, url)
			}
		} else if verbose {
			if dev {
				fmt.Printf("[INFO] DevDependency '%s' exists on npm (%s)\n", dep, url)
			} else {
				fmt.Printf("[INFO] Dependency '%s' exists on npm (%s)\n", dep, url)
			}
		}
	}

	// process dependencies
	for dep := range pkg.Dependencies {
		depWg.Add(1)
		go checkDep(dep, false)
	}

	// process devDependencies
	for dep := range pkg.DevDependencies {
		depWg.Add(1)
		go checkDep(dep, true)
	}

	depWg.Wait()
}

func main() {
	concurrency := flag.Int("c", 5, "Number of concurrent workers")
	verbose := flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	// read URLs from stdin
	scanner := bufio.NewScanner(os.Stdin)
	urls := make([]string, 0)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[ERROR] Failed to read input: %s\n", err)
		os.Exit(1)
	}

	// process URLs with a worker pool
	wg := &sync.WaitGroup{}
	semaphore := make(chan struct{}, *concurrency)

	for _, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(url string) {
			defer func() { <-semaphore }()
			processURL(url, *verbose, wg)
		}(url)
	}

	wg.Wait()
}
