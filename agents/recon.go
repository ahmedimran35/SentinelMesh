package agents

import (
	"context"
	"fmt"
	"strings"

	"sentinelmesh/fetchers"
	"sentinelmesh/models"
)

type ReconAgent struct {
	dns       *fetchers.DNSFetcher
	crtsh     *fetchers.CrtShFetcher
	internetdb *fetchers.InternetDBFetcher
	ipwhois   *fetchers.IPWhoisFetcher
	asn       *fetchers.ASNFetcher
}

func NewReconAgent(rateLimit int) *ReconAgent {
	return &ReconAgent{
		dns:       fetchers.NewDNSFetcher(rateLimit),
		crtsh:     fetchers.NewCrtShFetcher(rateLimit),
		internetdb: fetchers.NewInternetDBFetcher(rateLimit),
		ipwhois:   fetchers.NewIPWhoisFetcher(rateLimit),
		asn:       fetchers.NewASNFetcher(rateLimit),
	}
}

func (a *ReconAgent) Name() string        { return "recon" }
func (a *ReconAgent) Description() string { return "DNS, subdomain, port, and infrastructure reconnaissance" }

func (a *ReconAgent) Investigate(ctx context.Context, target models.Target, findings chan<- models.Finding) error {
	targetValue := target.Value
	var allIPs []string

	dnsResult, err := a.dns.Lookup(ctx, targetValue)
	if err == nil {
		details := fmt.Sprintf("A: %v\nAAAA: %v\nMX: %v\nNS: %v\nTXT: %v",
			dnsResult.A, dnsResult.AAAA, dnsResult.MX, dnsResult.NS, dnsResult.TXT)
		findings <- NewFinding("recon", "dns", models.RiskInfo, "DNS Records for "+targetValue, details, dnsResult)

		allIPs = dnsResult.GetIPs()

		if len(dnsResult.NS) > 0 {
			findings <- NewFinding("recon", "dns", models.RiskInfo, "Name Servers", strings.Join(dnsResult.NS, ", "), dnsResult.NS)
		}
	}

	crtResult, err := a.crtsh.Lookup(ctx, targetValue)
	if err == nil && len(crtResult.Subdomains) > 0 {
		details := fmt.Sprintf("Found %d subdomains:\n%s", len(crtResult.Subdomains), strings.Join(crtResult.Subdomains[:min(len(crtResult.Subdomains), 20)], "\n"))
		findings <- NewFinding("recon", "subdomain", models.RiskInfo,
			fmt.Sprintf("Subdomain Enumeration: %d found", len(crtResult.Subdomains)),
			details, crtResult.Subdomains)
	}

	for _, ip := range allIPs {
		dbResult, err := a.internetdb.Lookup(ctx, ip)
		if err == nil {
			if len(dbResult.Ports) > 0 {
				portStrs := make([]string, len(dbResult.Ports))
				for i, p := range dbResult.Ports {
					portStrs[i] = fmt.Sprintf("%d", p)
				}
				severity := models.RiskInfo
				if containsDangerousPort(dbResult.Ports) {
					severity = models.RiskHigh
				}
				findings <- NewFinding("recon", "port", severity,
					fmt.Sprintf("Open Ports on %s", ip),
					fmt.Sprintf("Ports: %s\nHostnames: %s\nCPEs: %s\nTags: %s",
						strings.Join(portStrs, ", "),
						strings.Join(dbResult.Hostnames, ", "),
						strings.Join(dbResult.CPES, ", "),
						strings.Join(dbResult.Tags, ", ")),
					dbResult)
			}

			// Vulnerabilities from InternetDB
			for _, vuln := range dbResult.Vulns {
				findings <- NewFinding("recon", "vuln", models.RiskHigh,
					fmt.Sprintf("Known Vulnerability on %s: %s", ip, vuln),
					fmt.Sprintf("InternetDB reports %s is vulnerable to %s", ip, vuln),
					vuln)
			}
		}

		// Geo lookup (free, no API key)
		geoResult, err := a.ipwhois.Lookup(ctx, ip)
		if err == nil {
			details := fmt.Sprintf("Country: %s\nCity: %s\nISP: %s\nOrg: %s\nASN: %s\nTimezone: %s\nProxy: %v\nHosting: %v",
				geoResult.Country, geoResult.City, geoResult.ISP, geoResult.Org, geoResult.ASN, geoResult.Timezone, geoResult.IsProxy, geoResult.IsHosting)
			findings <- NewFinding("recon", "geo", models.RiskInfo,
				fmt.Sprintf("Geolocation: %s (%s)", ip, geoResult.Country),
				details, geoResult)
		}

		// ASN lookup (free BGPView API, no key required)
		bgpResult, err := a.asn.LookupIP(ctx, ip)
		if err == nil && bgpResult.Data.ASN > 0 {
			asnNum := bgpResult.Data.ASN
			asnInfo := fmt.Sprintf("ASN: AS%d\nName: %s\nCountry: %s\nPrefix: %s\nRIR: %s",
				asnNum, bgpResult.Data.ASNName, bgpResult.Data.ASNCountry,
				bgpResult.Data.Prefix, bgpResult.Data.AllocationRIR)

			findings <- NewFinding("recon", "asn", models.RiskInfo,
				fmt.Sprintf("AS%d — %s", asnNum, bgpResult.Data.ASNName),
				asnInfo, bgpResult.Data)

			// Get prefix ranges
			prefixes, err := a.asn.GetPrefixes(ctx, asnNum)
			if err == nil && len(prefixes) > 0 {
				prefixStr := make([]string, min(len(prefixes), 10))
				for i, p := range prefixes {
					if i >= 10 {
						break
					}
					prefixStr[i] = p
				}
				findings <- NewFinding("recon", "prefix", models.RiskInfo,
					fmt.Sprintf("AS%d Prefixes: %d ranges", asnNum, len(prefixes)),
					fmt.Sprintf("Prefixes (%d total):\n%s", len(prefixes),
						strings.Join(prefixStr, "\n")), prefixes)
			}
		}
	}

	return nil
}

func containsDangerousPort(ports []int) bool {
	dangerous := map[int]bool{21: true, 23: true, 445: true, 3389: true, 5900: true, 6379: true, 27017: true}
	for _, p := range ports {
		if dangerous[p] {
			return true
		}
	}
	return false
}

// min uses Go builtin (1.22+)
