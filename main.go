package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"math"
	"time"
)

var lastIP string // Global variable to store the last known IP address

func main() {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	dnsName := os.Getenv("CLOUDFLARE_DNS_NAME")
	updateInterval := os.Getenv("CLOUDFLARE_DNS_UPDATE_INTERVAL")
	autoCreateDNS := os.Getenv("CLOUDFLARE_AUTO_CREATE_DNS") == "true"

	// Convert UPDATE_INTERVAL to an integer
	interval, err := strconv.Atoi(updateInterval)
	if err != nil || interval <= 0 {
		fmt.Printf("Invalid or missing UPDATE_INTERVAL '%s', defaulting to 600 seconds.\n", updateInterval)
		interval = 600 // Default to 600 seconds (10 minutes) if not set or invalid
	}

	ipVersion := os.Getenv("CLOUDFLARE_IP_VERSION")
	if ipVersion == "" {
		ipVersion = "ipv4"
	}

	ttl, err := strconv.Atoi(os.Getenv("CLOUDFLARE_DNS_TTL"))
	if err != nil || ttl <= 0 {
		ttl = 120
	}

	fmt.Printf("CloudFlare DDNS Updater started.\n")
	fmt.Printf("DNS Name: %s\n", dnsName)
	fmt.Printf("Update Interval: %d seconds\n", interval)
	fmt.Printf("Auto Create DNS: %t\n", autoCreateDNS)
	fmt.Printf("IP Version: %s\n", ipVersion)
	fmt.Printf("TTL: %d\n", ttl)

	// Get Zone ID and DNS Record ID
	zoneID, zoneName, err := getZoneID(apiToken, dnsName)
	if err != nil {
		fmt.Printf("Error getting Zone ID: %v\n", err)
		return
	}
	fmt.Printf("Found Zone: %s (ID: %s)\n", zoneName, zoneID)

	dnsRecordID, dnsRecordName, dnsRecordContent, err := getDNSRecordID(apiToken, zoneID, dnsName)
	if err != nil {
		fmt.Printf("Error getting DNS Record ID: %v\n", err)
		return
	}

	if dnsRecordID == "" {
		if autoCreateDNS {
			fmt.Printf("DNS record not found. Creating new record: %s\n", dnsName)
			ip, err := getPublicIP(ipVersion)
			if err != nil {
				fmt.Printf("Error getting public IP: %v\n", err)
				return
			}
			dnsRecordID, err = createDNSRecord(apiToken, zoneID, dnsName, ip, ipVersion, ttl)
			if err != nil {
				fmt.Printf("Error creating DNS record: %v\n", err)
				return
			}
			fmt.Printf("DNS record created successfully. ID: %s\n", dnsRecordID)
			lastIP = ip
		} else {
			fmt.Printf("DNS record not found and auto-creation is disabled. Exiting...\n")
			return
		}
	} else {
		fmt.Printf("Found DNS Record: %s (ID: %s, Content: %s)\n", dnsRecordName, dnsRecordID, dnsRecordContent)
		lastIP = dnsRecordContent
	}

	for {
		ip, err := getPublicIP(ipVersion)
		if err != nil {
			fmt.Printf("Error getting public IP: %v\n", err)
		} else {
			// Update the DNS record only if the IP address has changed
			if ip != lastIP {
				fmt.Printf("IP address changed. Old IP: %s, New IP: %s\n", lastIP, ip)
				if err := updateDNSRecord(apiToken, zoneID, dnsRecordID, dnsName, ip, ipVersion, ttl); err != nil {
					fmt.Printf("Error updating DNS record: %v\n", err)
				} else {
					fmt.Printf("DNS record updated successfully. %s -> %s\n", dnsName, ip)
					lastIP = ip // Update the last known IP address
				}
			} else {
				fmt.Printf("IP address unchanged (%s). No update necessary.\n", ip)
			}
		}

		fmt.Printf("Next update in %d seconds.\n", interval)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func getPublicIP(ipVersion string) (string, error) {
	var publicIPURL string
	if ipVersion == "ipv6" {
		publicIPURL = os.Getenv("CLOUDFLARE_IPV6_URL")
		if publicIPURL == "" {
			publicIPURL = "https://api64.ipify.org"
		}
	} else {
		publicIPURL = os.Getenv("CLOUDFLARE_IPV4_URL")
		if publicIPURL == "" {
			publicIPURL = "https://api.ipify.org"
		}
	}

	resp, err := http.Get(publicIPURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func extractDomain(dnsName string) string {
	parts := strings.Split(dnsName, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return dnsName
}

func getZoneID(apiToken, dnsName string) (string, string, error) {
	domain := extractDomain(dnsName)
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", domain)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := sendHTTPRequest(client, req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Result) > 0 {
		zoneID := result.Result[0].ID
		zoneName := result.Result[0].Name
		return zoneID, zoneName, nil
	}

	return "", "", fmt.Errorf("Zone not found for domain: %s", domain)
}

func getDNSRecordID(apiToken, zoneID, dnsName string) (string, string, string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s", zoneID, dnsName)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := sendHTTPRequest(client, req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Content string `json:"content"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Result) > 0 {
		dnsRecordID := result.Result[0].ID
		dnsRecordName := result.Result[0].Name
		dnsRecordContent := result.Result[0].Content
		return dnsRecordID, dnsRecordName, dnsRecordContent, nil
	}

	return "", dnsName, "", nil
}

func createDNSRecord(apiToken, zoneID, dnsName, ip string, ipVersion string, ttl int) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)
	data := map[string]interface{}{
		"type":    "A",
		"name":    dnsName,
		"content": ip,
		"ttl":     ttl,
	}
	if ipVersion == "ipv6" {
		data["type"] = "AAAA"
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := sendHTTPRequest(client, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Result.ID, nil
}

func updateDNSRecord(apiToken, zoneID, dnsRecordID, dnsName, ip string, ipVersion string, ttl int) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, dnsRecordID)
	data := map[string]interface{}{
		"type":    "A",
		"name":    dnsName,
		"content": ip,
		"ttl":     ttl,
	}
	if ipVersion == "ipv6" {
		data["type"] = "AAAA"
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := sendHTTPRequest(client, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func sendHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	const (
		maxRetries    = 3
		initialDelay  = 1 * time.Second
		maxDelay      = 30 * time.Second
		delayFactor   = 2.0
		jitterFactor  = 0.1
	)

	delay := initialDelay
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		fmt.Printf("Request failed (attempt %d/%d): %v\n", attempt, maxRetries, err)

		if attempt < maxRetries {
			jitter := time.Duration(math.Round(float64(delay) * jitterFactor))
			delay = time.Duration(float64(delay) * delayFactor)
			if delay > maxDelay {
				delay = maxDelay
			}

			fmt.Printf("Retrying in %v...\n", delay)
			time.Sleep(delay + jitter)
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts", maxRetries)
}