package runner

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Fixture struct {
	ProviderID string        `json:"provider_id"`
	Cases      []FixtureCase `json:"cases"`
}

type FixtureCase struct {
	Name          string            `json:"name"`
	Secret        string            `json:"secret"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	URL           string            `json:"url"`
	Params        map[string]string `json:"params"`
	Timestamp     int64             `json:"timestamp"`
	ExpectedError *string           `json:"expected_error"`
}

type ProviderSpec struct {
	ProviderID            string   `yaml:"provider_id"`
	Algorithm             string   `yaml:"algorithm"`
	SignatureHeader       string   `yaml:"signature_header"`
	SignaturePrefix       *string  `yaml:"signature_prefix"`
	SignatureEncoding     string   `yaml:"signature_encoding"`
	TimestampHeader       *string  `yaml:"timestamp_header"`
	PayloadConstruction   string   `yaml:"payload_construction"`
	PayloadTemplate       *string  `yaml:"payload_template"`
	ReplayWindowSeconds   *int     `yaml:"replay_window_seconds"`
	SignatureParsePattern *string  `yaml:"signature_parse_pattern"`
	TimestampParsePattern *string  `yaml:"timestamp_parse_pattern"`
	MultipleSignatures    bool     `yaml:"multiple_signatures"`
}

func RunFixtureFile(path string) (int, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return 0, 0, err
	}
	if fixture.ProviderID == "" {
		return 0, 0, fmt.Errorf("missing provider_id")
	}

	// Load provider spec
	specPath := filepath.Join(filepath.Dir(filepath.Dir(path)), "providers", fixture.ProviderID+".yaml")
	spec, err := loadProviderSpec(specPath)
	if err != nil {
		return 0, 0, fmt.Errorf("load provider spec: %w", err)
	}

	pass, fail := 0, 0
	for _, c := range fixture.Cases {
		if c.Name == "" {
			fail++
			continue
		}
		err := verifyFixtureCase(spec, c)
		if c.ExpectedError == nil {
			// Should pass
			if err == nil {
				pass++
			} else {
				fail++
			}
		} else {
			// Should fail with specific error
			if err != nil && strings.Contains(err.Error(), *c.ExpectedError) {
				pass++
			} else {
				fail++
			}
		}
	}
	return pass, fail, nil
}

func loadProviderSpec(path string) (*ProviderSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec ProviderSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func verifyFixtureCase(spec *ProviderSpec, c FixtureCase) error {
	// Check signature header exists
	sigHeader, ok := c.Headers[spec.SignatureHeader]
	if !ok {
		return fmt.Errorf("ERR_MISSING_SIGNATURE")
	}

	// Extract signature(s)
	signatures := extractSignatures(spec, sigHeader)
	if len(signatures) == 0 {
		return fmt.Errorf("ERR_BAD_FORMAT")
	}

	// Extract timestamp if needed
	var timestamp int64
	if spec.TimestampHeader != nil && *spec.TimestampHeader != "" {
		tsHeader, ok := c.Headers[*spec.TimestampHeader]
		if !ok {
			return fmt.Errorf("ERR_MISSING_TIMESTAMP")
		}
		ts, err := strconv.ParseInt(tsHeader, 10, 64)
		if err != nil {
			return fmt.Errorf("ERR_BAD_FORMAT")
		}
		timestamp = ts
	} else if spec.TimestampParsePattern != nil {
		// Extract timestamp from signature header (e.g., Stripe)
		re := regexp.MustCompile(*spec.TimestampParsePattern)
		matches := re.FindStringSubmatch(sigHeader)
		if len(matches) < 2 {
			return fmt.Errorf("ERR_BAD_FORMAT")
		}
		ts, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return fmt.Errorf("ERR_BAD_FORMAT")
		}
		timestamp = ts
	} else {
		timestamp = c.Timestamp
	}

	// Check replay window
	if spec.ReplayWindowSeconds != nil && *spec.ReplayWindowSeconds > 0 {
		if c.Timestamp > 0 && timestamp > 0 {
			diff := c.Timestamp - timestamp
			if diff < 0 {
				diff = -diff
			}
			if diff > int64(*spec.ReplayWindowSeconds) {
				return fmt.Errorf("ERR_TIMESTAMP_EXPIRED")
			}
		}
	}

	// Construct payload to sign
	payload := constructPayload(spec, c, timestamp)

	// Compute expected signature
	expectedSig, err := computeSignature(spec, c.Secret, payload)
	if err != nil {
		return fmt.Errorf("ERR_COMPUTE: %w", err)
	}

	// Compare signatures (timing-safe)
	valid := false
	for _, sig := range signatures {
		if timingSafeEqual(sig, expectedSig) {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("ERR_BAD_SIGNATURE")
	}

	return nil
}

func extractSignatures(spec *ProviderSpec, header string) []string {
	var sigs []string

	if spec.SignatureParsePattern != nil {
		// Extract using regex (e.g., Stripe v1=...)
		re := regexp.MustCompile(*spec.SignatureParsePattern)
		matches := re.FindAllStringSubmatch(header, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				sigs = append(sigs, match[1])
			}
		}
	} else if spec.SignaturePrefix != nil && *spec.SignaturePrefix != "" {
		// Remove prefix (e.g., "sha256=", "v0=")
		if strings.HasPrefix(header, *spec.SignaturePrefix) {
			sigs = append(sigs, strings.TrimPrefix(header, *spec.SignaturePrefix))
		} else {
			return nil // Prefix mismatch
		}
	} else {
		// Bare signature
		sigs = append(sigs, header)
	}

	return sigs
}

func constructPayload(spec *ProviderSpec, c FixtureCase, timestamp int64) string {
	switch spec.PayloadConstruction {
	case "raw_body":
		return c.Body
	case "custom":
		if spec.PayloadTemplate == nil {
			return c.Body
		}
		template := *spec.PayloadTemplate
		template = strings.ReplaceAll(template, "{timestamp}", strconv.FormatInt(timestamp, 10))
		template = strings.ReplaceAll(template, "{body}", c.Body)
		template = strings.ReplaceAll(template, "{url}", c.URL)
		if strings.Contains(template, "{sorted_params}") {
			sortedParams := sortParams(c.Params)
			template = strings.ReplaceAll(template, "{sorted_params}", sortedParams)
		}
		return template
	default:
		return c.Body
	}
}

func sortParams(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+params[k])
	}
	return strings.Join(parts, "")
}

func computeSignature(spec *ProviderSpec, secret, payload string) (string, error) {
	var mac []byte
	switch spec.Algorithm {
	case "hmac-sha256":
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(payload))
		mac = h.Sum(nil)
	case "hmac-sha1":
		h := hmac.New(sha1.New, []byte(secret))
		h.Write([]byte(payload))
		mac = h.Sum(nil)
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", spec.Algorithm)
	}

	switch spec.SignatureEncoding {
	case "hex":
		return hex.EncodeToString(mac), nil
	case "base64":
		return base64.StdEncoding.EncodeToString(mac), nil
	default:
		return "", fmt.Errorf("unsupported encoding: %s", spec.SignatureEncoding)
	}
}

func timingSafeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func FixturePath(root, provider string) string {
	return filepath.Join(root, "fixtures", provider+".fixtures.json")
}