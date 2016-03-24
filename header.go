package hstspreload

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Value to use for maxAgeSeconds when invalid.
	// Note that the struct will still be inialized with a value of
	// maxAgeSeconds == 0, so this value is only used as a best effort to
	// catch bugs.
	BOGUS_MAX_AGE = (1<<64 - 1)
)

type HSTSHeader struct {
	preload           bool
	includeSubDomains bool
	maxAgePresent     bool
	// maxAgeSeconds is valid iff maxAgePreset == true
	// It is recommended to set this to BOGUS_MAX_AGE when invalid.
	maxAgeSeconds uint64
}

// Mainly useful for testing.
func headersEqual(header1 HSTSHeader, header2 HSTSHeader) bool {
	if header1.preload != header2.preload {
		return false
	}

	if header1.includeSubDomains != header2.includeSubDomains {
		return false
	}

	if header1.maxAgePresent != header2.maxAgePresent {
		return false
	}

	if header1.maxAgePresent && (header1.maxAgeSeconds != header2.maxAgeSeconds) {
		return false
	}

	return true
}

func ParseHeaderString(headerString string) (HSTSHeader, error) {
	var hstsHeader HSTSHeader

	hstsHeader.preload = false
	hstsHeader.includeSubDomains = false
	hstsHeader.maxAgePresent = false
	hstsHeader.maxAgeSeconds = BOGUS_MAX_AGE

	hstsParts := strings.Split(headerString, ";")
	for i, part := range hstsParts {
		// TODO: Trim tabs? https://crbug.com/596561#c10
		hstsParts[i] = strings.TrimSpace(part)
	}

	for _, part := range hstsParts {
		part = strings.ToLower(part)
		switch {
		case part == "preload":
			hstsHeader.preload = true

		case part == "includesubdomains":
			hstsHeader.includeSubDomains = true

		case strings.HasPrefix(part, "max-age="):
			maxAgeNumericalString := part[8:]
			// TODO the numerical string contains only digits, no symbols (no "+")
			maxAge, err := strconv.ParseUint(maxAgeNumericalString, 10, 63)
			if err != nil {
				return hstsHeader, fmt.Errorf("Could not parse max-age value [%s].", maxAgeNumericalString)
			}
			hstsHeader.maxAgePresent = true
			hstsHeader.maxAgeSeconds = maxAge

		case strings.HasPrefix(part, "max-age"):
			return hstsHeader, fmt.Errorf("The max-age directive name is present without a value.")

			// TODO: warn on unknown directives/tokens.
		}
	}

	// TODO: show multiple errors
	// TODO: warn on repeated directives
	// TODO: Warn when max-age is longer than a year (Chrome will truncate)
	// TODO: Allow testing whether the header is valid according to the spec (vs. having all preload requirements)
	// TODO: warn on empty directives / extra semicolons

	return hstsHeader, nil
}

func CheckHeader(hstsHeader HSTSHeader) error {
	// TODO: display all error, e.g. missing preload as well as includeSubDomains
	// TODO: Is it idiomatic to return only an error, with nil meaning success? (same goes for other functions in hstspreload)

	if !hstsHeader.includeSubDomains {
		return fmt.Errorf("Must have the `includeSubDomains` directive.")
	}

	if !hstsHeader.preload {
		return fmt.Errorf("Must have the `preload` directive.")
	}

	if !hstsHeader.maxAgePresent {
		return fmt.Errorf("Must have the `max-age` directive.")
	}

	if hstsHeader.maxAgeSeconds < 10886400 {
		return fmt.Errorf("The max-age must be at least 10886400 seconds (== 18 weeks). The header had max-age=%d", hstsHeader.maxAgeSeconds)
	}

	return nil
}

func CheckHeaderString(headerString string) error {
	hstsHeader, err := ParseHeaderString(headerString)

	if err != nil {
		return fmt.Errorf("Error parsing HSTS header.")
	}

	return CheckHeader(hstsHeader)
}
