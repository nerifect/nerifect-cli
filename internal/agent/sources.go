package agent

import (
	"github.com/nerifect/nerifect-cli/internal/store"
)

// DefaultSources are well-known compliance document URLs seeded on first daemon start.
var DefaultSources = []struct {
	URL  string
	Name string
}{
	{"https://owasp.org/Top10/", "OWASP Top 10"},
	{"https://gdpr-info.eu/", "GDPR"},
	{"https://www.pcisecuritystandards.org/document_library/", "PCI DSS"},
	{"https://artificialintelligenceact.eu/the-act/", "EU AI Act"},
	{"https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final", "NIST 800-53"},
}

// SeedDefaultSources inserts default sources if the agent_sources table is empty.
func SeedDefaultSources() error {
	count, err := store.AgentSourceCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, s := range DefaultSources {
		if _, err := store.CreateAgentSource(s.URL, s.Name); err != nil {
			return err
		}
	}
	return nil
}
