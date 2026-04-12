package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

// ============================================================================
// Module: PII Masking Scanner (V12-D Extension)
// Scans neuron content for sensitive data before cartridge export.
// Masks emails, API keys, tokens, passwords, IPs, phone numbers.
// Must run BEFORE encryption to ensure clean cartridge for marketplace.
// ============================================================================

// PIIMaskResult holds the masking statistics
type PIIMaskResult struct {
	EmailsFound   int
	KeysFound     int
	IPsFound      int
	PhonesFound   int
	PasswordFound int
	TotalMasked   int
}

// piiPatterns defines regex patterns for common sensitive data
var piiPatterns = []struct {
	Name    string
	Pattern *regexp.Regexp
	Replace string
}{
	{
		Name:    "email",
		Pattern: regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
		Replace: "[EMAIL_REDACTED]",
	},
	{
		Name:    "api_key",
		Pattern: regexp.MustCompile(`(?i)(api[_\-]?key|apikey|api_secret|secret_key|access_key)\s*[:=]\s*["']?([a-zA-Z0-9\-_]{16,})["']?`),
		Replace: "${1}: [KEY_REDACTED]",
	},
	{
		Name:    "bearer_token",
		Pattern: regexp.MustCompile(`(?i)(bearer|token|authorization)\s*[:=]\s*["']?([a-zA-Z0-9\-_.]{20,})["']?`),
		Replace: "${1}: [TOKEN_REDACTED]",
	},
	{
		Name:    "password",
		Pattern: regexp.MustCompile(`(?i)(password|passwd|pwd|pass)\s*[:=]\s*["']?([^\s"']{4,})["']?`),
		Replace: "${1}: [PASSWORD_REDACTED]",
	},
	{
		Name:    "ipv4",
		Pattern: regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`),
		Replace: "[IP_REDACTED]",
	},
	{
		Name:    "phone_kr",
		Pattern: regexp.MustCompile(`\b(01[016789]-?\d{3,4}-?\d{4})\b`),
		Replace: "[PHONE_REDACTED]",
	},
	{
		Name:    "phone_intl",
		Pattern: regexp.MustCompile(`\+\d{1,3}[\s\-]?\d{2,4}[\s\-]?\d{3,4}[\s\-]?\d{3,4}`),
		Replace: "[PHONE_REDACTED]",
	},
	{
		Name:    "telegram_token",
		Pattern: regexp.MustCompile(`\d{8,10}:[A-Za-z0-9_\-]{35,}`),
		Replace: "[TELEGRAM_TOKEN_REDACTED]",
	},
	{
		Name:    "aws_key",
		Pattern: regexp.MustCompile(`(?i)(AKIA|ABIA|ACCA|ASIA)[A-Z0-9]{16}`),
		Replace: "[AWS_KEY_REDACTED]",
	},
	{
		Name:    "private_key",
		Pattern: regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE KEY-----`),
		Replace: "[PRIVATE_KEY_REDACTED]",
	},
}

// whitelistIPs are IPs that should NOT be masked (localhost, etc)
var whitelistIPs = map[string]bool{
	"127.0.0.1": true,
	"0.0.0.0":   true,
	"localhost":  true,
}

// MaskPII scans content and replaces all detected PII with safe placeholders
func MaskPII(content string) (string, PIIMaskResult) {
	result := PIIMaskResult{}
	masked := content

	for _, p := range piiPatterns {
		matches := p.Pattern.FindAllString(masked, -1)
		if len(matches) == 0 {
			continue
		}

		// Special handling: whitelist certain IPs
		if p.Name == "ipv4" {
			for _, m := range matches {
				if whitelistIPs[m] {
					continue
				}
				masked = strings.Replace(masked, m, p.Replace, 1)
				result.IPsFound++
				result.TotalMasked++
			}
			continue
		}

		count := len(matches)
		masked = p.Pattern.ReplaceAllString(masked, p.Replace)

		switch p.Name {
		case "email":
			result.EmailsFound += count
		case "api_key", "bearer_token", "telegram_token", "aws_key", "private_key":
			result.KeysFound += count
		case "password":
			result.PasswordFound += count
		case "phone_kr", "phone_intl":
			result.PhonesFound += count
		}
		result.TotalMasked += count
	}

	return masked, result
}

// MaskPIISummary returns a human-readable summary of what was masked
func MaskPIISummary(r PIIMaskResult) string {
	if r.TotalMasked == 0 {
		return "No PII detected"
	}
	parts := []string{}
	if r.EmailsFound > 0 {
		parts = append(parts, fmt.Sprintf("%d emails", r.EmailsFound))
	}
	if r.KeysFound > 0 {
		parts = append(parts, fmt.Sprintf("%d keys/tokens", r.KeysFound))
	}
	if r.PasswordFound > 0 {
		parts = append(parts, fmt.Sprintf("%d passwords", r.PasswordFound))
	}
	if r.IPsFound > 0 {
		parts = append(parts, fmt.Sprintf("%d IPs", r.IPsFound))
	}
	if r.PhonesFound > 0 {
		parts = append(parts, fmt.Sprintf("%d phones", r.PhonesFound))
	}
	return fmt.Sprintf("Masked %d items: %s", r.TotalMasked, strings.Join(parts, ", "))
}

// maskTarGzPII processes a tar.gz byte slice in-memory,
// masking PII in all text-based files (.neuron, .md, .txt, .jsonl)
func maskTarGzPII(tarGzData []byte) ([]byte, error) {
	// Decompress
	gzReader, err := gzip.NewReader(strings.NewReader(string(tarGzData)))
	if err != nil {
		return tarGzData, err
	}
	defer gzReader.Close()

	// Read all entries
	tarReader := tar.NewReader(gzReader)
	type entry struct {
		Header *tar.Header
		Data   []byte
	}
	var entries []entry
	totalMasked := PIIMaskResult{}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return tarGzData, err
		}

		headerCopy := *header
		var data []byte
		if header.Typeflag == tar.TypeReg {
			data, _ = io.ReadAll(tarReader)

			// Mask PII in text-based files
			ext := strings.ToLower(filepath.Ext(header.Name))
			if ext == ".neuron" || ext == ".md" || ext == ".txt" || ext == ".jsonl" {
				masked, result := MaskPII(string(data))
				if result.TotalMasked > 0 {
					data = []byte(masked)
					headerCopy.Size = int64(len(data))
					totalMasked.TotalMasked += result.TotalMasked
					totalMasked.EmailsFound += result.EmailsFound
					totalMasked.KeysFound += result.KeysFound
					totalMasked.PasswordFound += result.PasswordFound
					totalMasked.IPsFound += result.IPsFound
					totalMasked.PhonesFound += result.PhonesFound
				}
			}
		}
		entries = append(entries, entry{Header: &headerCopy, Data: data})
	}

	// Recompress with masked content
	outBuf := &bytesBuffer{}
	gzWriter := gzip.NewWriter(outBuf)
	tarWriter := tar.NewWriter(gzWriter)

	for _, e := range entries {
		_ = tarWriter.WriteHeader(e.Header)
		if len(e.Data) > 0 {
			tarWriter.Write(e.Data)
		}
	}

	tarWriter.Close()
	gzWriter.Close()

	fmt.Printf("[PII Masker] %s\n", MaskPIISummary(totalMasked))
	return outBuf.Bytes(), nil
}

