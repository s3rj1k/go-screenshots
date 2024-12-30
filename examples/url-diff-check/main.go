package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/png" // importing PNG decoder
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/corona10/goimagehash"
	"golang.org/x/net/publicsuffix"

	screenshot "github.com/s3rj1k/go-webpage-screenshots"
)

func computeImageDifferenceDistance(left, right []byte) (int, error) {
	leftImage, _, err := image.Decode(bytes.NewReader(left))
	if err != nil {
		return 0, err
	}

	rightImage, _, err := image.Decode(bytes.NewReader(right))
	if err != nil {
		return 0, err
	}

	leftImageHash, err := goimagehash.DifferenceHash(leftImage)
	if err != nil {
		return 0, err
	}

	rightImageHash, err := goimagehash.DifferenceHash(rightImage)
	if err != nil {
		return 0, err
	}

	distance, err := leftImageHash.Distance(rightImageHash)
	if err != nil {
		return 0, err
	}

	return distance, nil
}

func getSubDomainPrefix(domain, gtldPlusOne string) string {
	prefix := strings.TrimSuffix(domain, gtldPlusOne)
	if len(prefix) > 0 && strings.HasSuffix(prefix, ".") && prefix != "." {
		return strings.TrimSuffix(prefix, ".")
	}

	return "_"
}

func createScreenshotOrReportImageDifferenceDistance(rawURL string) (int, error) {
	urlObj, err := url.Parse(rawURL)
	if err != nil {
		return 0, err
	}

	gtldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(urlObj.Host)
	if err != nil {
		return 0, err
	}

	encodedURL := base64.URLEncoding.EncodeToString([]byte(urlObj.String()))
	subdomain := getSubDomainPrefix(urlObj.Host, gtldPlusOne)
	imagePath := filepath.Join(
		gtldPlusOne, subdomain, fmt.Sprintf("%s.png", encodedURL),
	)

	config := screenshot.DefaultConfig()
	config.URL = urlObj.String()
	config.FullPage = true

	newImage, err := config.Screenshot()
	if err != nil {
		return 0, err
	}

	err = os.MkdirAll(filepath.Dir(imagePath), 0755)
	if err != nil {
		return 0, err
	}

	existingImage, err := os.ReadFile(imagePath)
	if errors.Is(err, os.ErrNotExist) {
		return 0, os.WriteFile(imagePath, newImage, 0644)
	}
	if err != nil {
		return 0, err
	}

	return computeImageDifferenceDistance(existingImage, newImage)
}

func main() {
	url := os.Getenv("URL")
	if url == "" {
		fmt.Fprintf(os.Stderr, "URL environment variable is required\n")
		os.Exit(1)
	}

	distance, err := createScreenshotOrReportImageDifferenceDistance(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if distance == 0 {
		fmt.Println("New screenshot saved")
	} else {
		fmt.Printf("Image difference distance: %d\n", distance)
	}
}
