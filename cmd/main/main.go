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
	"github.com/gkettani/bobber-the-swe/internal/services/web"
)

func main() {
	logger.Info("Starting job processing application with web interface")

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

	// Create web service
	webService := web.NewWebService(orchestrator)

	// Start the pipeline
	if err := orchestrator.Start(ctx); err != nil {
		logger.Error(fmt.Sprintf("Failed to start orchestrator: %v", err))
		panic(err)
	}

	// Start the web service
	webCtx, webCancel := context.WithCancel(ctx)
	go func() {
		if err := webService.Start(webCtx); err != nil {
			logger.Error("Web service error", "error", err)
		}
	}()

	// Log startup status
	status := orchestrator.GetStatus()
	logger.Info(fmt.Sprintf("Pipeline started successfully - Discovery companies: %d, Enrichment companies: %d, Queue size: %d",
		status.DiscoveryCompanies, status.EnrichmentCompanies, status.QueueSize))
	logger.Info(fmt.Sprintf("Pipeline uptime: %s", status.Uptime))
	logger.Info(fmt.Sprintf("Web interface available at: http://%s:%d", webService.GetHost(), webService.GetPort()))

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Received shutdown signal, stopping services gracefully...")

	// Log final metrics before shutdown
	finalMetrics := orchestrator.GetMetrics()
	logger.Info(fmt.Sprintf("Final metrics - Jobs processed: %d, Success rate: %.2f%%, Discovery cycles: %d",
		finalMetrics.JobsProcessed, finalMetrics.CalculateSuccessRate(), finalMetrics.DiscoveryCycles))

	// Stop services
	webCancel() // Stop web service
	if err := webService.Stop(); err != nil {
		logger.Error("Error stopping web service", "error", err)
	}

	if err := orchestrator.Stop(); err != nil {
		logger.Error("Error stopping orchestrator", "error", err)
	}

	// Cancel context to stop all workers
	cancel()

	logger.Info("Application shutdown complete")
}
