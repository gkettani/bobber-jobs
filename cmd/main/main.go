package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/services/deduplication"
	"github.com/gkettani/bobber-the-swe/internal/services/discovery"
	"github.com/gkettani/bobber-the-swe/internal/services/enrichment"
	"github.com/gkettani/bobber-the-swe/internal/services/orchestration"
	"github.com/gkettani/bobber-the-swe/internal/services/persistence"
)

func main() {
	logger.Info("Starting job processing application with new architecture")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	discoveryService, err := discovery.NewJobDiscoveryService("config/companies.yaml")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create discovery service: %v", err))
		panic(err)
	}

	enrichmentService, err := enrichment.NewJobEnrichmentService("config/scrapers.yaml")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create enrichment service: %v", err))
		panic(err)
	}

	persistenceService := persistence.NewJobPersistenceService(100)
	deduplicationService := deduplication.NewDeduplicationService()

	// Create orchestrator configuration
	config := orchestration.DefaultConfig()

	// Create and start orchestrator
	orchestrator := orchestration.NewOrchestrator(
		config,
		discoveryService,
		enrichmentService,
		persistenceService,
		deduplicationService,
	)

	// Start the pipeline
	if err := orchestrator.Start(ctx); err != nil {
		logger.Error(fmt.Sprintf("Failed to start orchestrator: %v", err))
		panic(err)
	}

	// Log startup status using structured models
	status := orchestrator.GetStatus()
	logger.Info(fmt.Sprintf("Pipeline started successfully - Discovery companies: %d, Enrichment companies: %d, Queue size: %d",
		status.DiscoveryCompanies, status.EnrichmentCompanies, status.QueueSize))
	logger.Info(fmt.Sprintf("Pipeline uptime: %s", status.Uptime))

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Received shutdown signal, stopping pipeline gracefully...")

	// Log final metrics before shutdown
	finalMetrics := orchestrator.GetMetrics()
	logger.Info(fmt.Sprintf("Final metrics - Jobs processed: %d, Success rate: %.2f%%, Discovery cycles: %d",
		finalMetrics.JobsProcessed, finalMetrics.CalculateSuccessRate(), finalMetrics.DiscoveryCycles))

	// Stop the orchestrator
	if err := orchestrator.Stop(); err != nil {
		logger.Error(fmt.Sprintf("Error stopping orchestrator: %v", err))
	}

	// Cancel context to stop all workers
	cancel()

	logger.Info("Application shutdown complete")
}
