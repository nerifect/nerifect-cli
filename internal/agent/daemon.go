package agent

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/llm"
	"github.com/nerifect/nerifect-cli/internal/policy"
	"github.com/nerifect/nerifect-cli/internal/store"
)

// RunDaemon is the main entry point for the background daemon process.
func RunDaemon(cfg *config.Config) error {
	logger := log.New(os.Stdout, "[agent] ", log.LstdFlags)
	logger.Println("Starting policy ingestion agent")

	// Open database.
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	// Seed default sources if table is empty.
	if err := SeedDefaultSources(); err != nil {
		logger.Printf("Warning: failed to seed default sources: %v", err)
	}

	// Build LLM client and policy manager.
	llmClient := llm.NewClient(cfg.LLMProvider, cfg.ActiveAPIKey(), cfg.DefaultModel)
	mgr := policy.NewManager(llmClient)
	fetcher := policy.NewFetcher()

	// Set up graceful shutdown.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Run initial check cycle.
	logger.Println("Running initial check cycle")
	runChecks(ctx, logger, fetcher, mgr)

	// Start ticker for periodic checks.
	interval := time.Duration(cfg.AgentCheckInterval) * time.Hour
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Printf("Scheduled checks every %d hours", cfg.AgentCheckInterval)

	for {
		select {
		case <-ctx.Done():
			logger.Println("Shutting down")
			// Clean up PID file.
			os.Remove(PidPath(cfg))
			return nil
		case <-ticker.C:
			logger.Println("Starting scheduled check cycle")
			runChecks(ctx, logger, fetcher, mgr)
		}
	}
}

// runChecks iterates all enabled sources and checks each one.
func runChecks(ctx context.Context, logger *log.Logger, fetcher *policy.Fetcher, mgr *policy.Manager) {
	sources, err := store.ListEnabledAgentSources()
	if err != nil {
		logger.Printf("Error listing sources: %v", err)
		return
	}

	if len(sources) == 0 {
		logger.Println("No enabled sources to check")
		return
	}

	logger.Printf("Checking %d source(s)", len(sources))
	for _, src := range sources {
		// Check for shutdown between sources.
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := checkSource(ctx, logger, fetcher, mgr, src); err != nil {
			logger.Printf("Error checking %q: %v", src.URL, err)
			_ = store.UpdateAgentSourceError(src.ID, err.Error())
		}
	}
}

// checkSource fetches a single source, compares content hash, and ingests if changed.
func checkSource(ctx context.Context, logger *log.Logger, fetcher *policy.Fetcher, mgr *policy.Manager, src store.AgentSource) error {
	logger.Printf("Fetching %s", src.URL)

	text, err := fetcher.FetchURL(src.URL)
	if err != nil {
		return fmt.Errorf("fetching: %w", err)
	}

	// Compute content hash.
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(text)))

	// If content unchanged, just update the check timestamp.
	if hash == src.ContentHash {
		logger.Printf("No changes detected for %s", src.URL)
		return store.UpdateAgentSourceCheck(src.ID, hash, src.LinkedPolicyID)
	}

	logger.Printf("Content changed for %s, ingesting policy", src.URL)

	// Delete old linked policy if it exists.
	if src.LinkedPolicyID > 0 {
		if err := store.DeletePolicy(src.LinkedPolicyID); err != nil {
			logger.Printf("Warning: could not delete old policy %d: %v", src.LinkedPolicyID, err)
		}
	}

	// Ingest the new content via LLM pipeline.
	p, err := mgr.AddFromText(ctx, text, src.URL)
	if err != nil {
		return fmt.Errorf("ingesting policy: %w", err)
	}

	logger.Printf("Ingested policy %q (ID %d) with %d rules", p.Name, p.ID, p.RuleCount)
	return store.UpdateAgentSourceCheck(src.ID, hash, p.ID)
}
