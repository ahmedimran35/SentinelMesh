package fetchers

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

type NVDFetcher struct {
	*BaseFetcher
}

type NVDResponse struct {
	ResultsPerPage int         `json:"resultsPerPage"`
	StartIndex     int         `json:"startIndex"`
	TotalResults   int         `json:"totalResults"`
	Vulnerabilities []NVDVuln `json:"vulnerabilities"`
}

type NVDVuln struct {
	CVE NVDCVE `json:"cve"`
}

type NVDCVE struct {
	ID               string              `json:"id"`
	SourceIdentifier string              `json:"sourceIdentifier"`
	Published        time.Time           `json:"published"`
	LastModified     time.Time           `json:"lastModified"`
	VulnStatus       string              `json:"vulnStatus"`
	Descriptions     []NVDDescription    `json:"descriptions"`
	Metrics          *NVDMetrics         `json:"metrics,omitempty"`
	Weaknesses       []NVDWeakness       `json:"weaknesses,omitempty"`
	References       []NVDReference      `json:"references,omitempty"`
}

type NVDDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type NVDMetrics struct {
	CvssMetricV31 []NVDCvssMetric `json:"cvssMetricV31,omitempty"`
	CvssMetricV2  []NVDCvssMetric `json:"cvssMetricV2,omitempty"`
}

type NVDCvssMetric struct {
	Source   string       `json:"source"`
	Type     string       `json:"type"`
	CvssData NVDCvssData  `json:"cvssData"`
}

type NVDCvssData struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	BaseScore             float64 `json:"baseScore"`
	BaseSeverity          string  `json:"baseSeverity"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
}

type NVDWeakness struct {
	Source string         `json:"source"`
	Type   string         `json:"type"`
	Desc   []NVDDescription `json:"description"`
}

type NVDReference struct {
	URL    string   `json:"url"`
	Source string   `json:"source"`
	Tags   []string `json:"tags,omitempty"`
}

type CVEResult struct {
	ID          string
	Description string
	BaseScore   float64
	Severity    string
	Published   time.Time
	References  []string
}

func NewNVDFetcher(rateLimit int) *NVDFetcher {
	return &NVDFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

// SearchCVEs searches for CVEs by keyword (software name, version, etc.)
func (f *NVDFetcher) SearchCVEs(ctx context.Context, keyword string, maxResults int) ([]CVEResult, error) {
	params := url.Values{}
	params.Set("keywordSearch", keyword)
	params.Set("resultsPerPage", fmt.Sprintf("%d", maxResults))

	reqURL := "https://services.nvd.nist.gov/rest/json/cves/2.0?" + params.Encode()

	var resp NVDResponse
	if err := f.GetJSON(ctx, reqURL, nil, &resp); err != nil {
		return nil, fmt.Errorf("NVD search for '%s' failed: %w", keyword, err)
	}

	var results []CVEResult
	for _, v := range resp.Vulnerabilities {
		cve := v.CVE
		result := CVEResult{
			ID:        cve.ID,
			Published: cve.Published,
		}

		// Get English description
		for _, d := range cve.Descriptions {
			if d.Lang == "en" {
				result.Description = d.Value
				break
			}
		}

		// Get CVSS score
		if cve.Metrics != nil {
			if len(cve.Metrics.CvssMetricV31) > 0 {
				result.BaseScore = cve.Metrics.CvssMetricV31[0].CvssData.BaseScore
				result.Severity = cve.Metrics.CvssMetricV31[0].CvssData.BaseSeverity
			} else if len(cve.Metrics.CvssMetricV2) > 0 {
				result.BaseScore = cve.Metrics.CvssMetricV2[0].CvssData.BaseScore
				// V2 uses numeric score ranges
				if result.BaseScore >= 7.0 {
					result.Severity = "HIGH"
				} else if result.BaseScore >= 4.0 {
					result.Severity = "MEDIUM"
				} else {
					result.Severity = "LOW"
				}
			}
		}

		// Get references
		for _, ref := range cve.References {
			result.References = append(result.References, ref.URL)
		}

		results = append(results, result)
	}

	return results, nil
}

// GetCVE retrieves a specific CVE by ID
func (f *NVDFetcher) GetCVE(ctx context.Context, cveID string) (*CVEResult, error) {
	reqURL := fmt.Sprintf("https://services.nvd.nist.gov/rest/json/cves/2.0?cveId=%s", cveID)

	var resp NVDResponse
	if err := f.GetJSON(ctx, reqURL, nil, &resp); err != nil {
		return nil, fmt.Errorf("NVD lookup for %s failed: %w", cveID, err)
	}

	if len(resp.Vulnerabilities) == 0 {
		return nil, fmt.Errorf("CVE %s not found", cveID)
	}

	cve := resp.Vulnerabilities[0].CVE
	result := &CVEResult{
		ID:        cve.ID,
		Published: cve.Published,
	}

	for _, d := range cve.Descriptions {
		if d.Lang == "en" {
			result.Description = d.Value
			break
		}
	}

	if cve.Metrics != nil {
		if len(cve.Metrics.CvssMetricV31) > 0 {
			result.BaseScore = cve.Metrics.CvssMetricV31[0].CvssData.BaseScore
			result.Severity = cve.Metrics.CvssMetricV31[0].CvssData.BaseSeverity
		} else if len(cve.Metrics.CvssMetricV2) > 0 {
			result.BaseScore = cve.Metrics.CvssMetricV2[0].CvssData.BaseScore
			if result.BaseScore >= 7.0 {
				result.Severity = "HIGH"
			} else if result.BaseScore >= 4.0 {
				result.Severity = "MEDIUM"
			} else {
				result.Severity = "LOW"
			}
		}
	}

	for _, ref := range cve.References {
		result.References = append(result.References, ref.URL)
	}

	return result, nil
}
