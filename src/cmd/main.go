package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/cache"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/logger"
	"github.com/brendank310/aztui/pkg/resourceviews"
	"github.com/brendank310/aztui/pkg/tracing"

	"github.com/gdamore/tcell/v2"
	_ "github.com/rivo/tview"
)

type AzTuiState struct {
	// Basic TUI variables
	*resourceviews.AppLayout
	config.Config
	CacheService *cache.ResourceCacheService
}

func NewAzTuiState() *AzTuiState {
	// Base initialization
	err := logger.InitLogger()
	if err != nil {
		panic(err)
	}

	configPath := os.Getenv("AZTUI_CONFIG_PATH")
	if configPath == "" {
		configPath = os.Getenv("HOME") + "/.config/aztui.yaml"
	}

	c, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	// Initialize cache service with configured TTL
	cacheService := cache.NewResourceCacheService(c.GetCacheTTL())

	a := AzTuiState{
		AppLayout:    resourceviews.NewAppLayout(),
		Config:       c,
		CacheService: cacheService,
	}

	// Initialize cache service globally for views to use
	resourceviews.SetCacheService(cacheService)

	subscriptionList := resourceviews.NewSubscriptionListView(a.AppLayout)
	if subscriptionList == nil {
		panic("unable to create a subscription list")
	}

	a.AppLayout.InputField.SetFinishedFunc(func(key tcell.Key) {
		if a.FocusedViewIndex == 0 {
			a.App.SetFocus(subscriptionList.List)
		} else if a.FocusedViewIndex == 1 {
			a.App.SetFocus(subscriptionList.ResourceGroupListView.List)
		}
	})

	return &a
}

func main() {
	// Parse command line flags
	var (
		perfTrace     = flag.Bool("perf-trace", false, "Enable performance tracing and output statistics to stderr")
		perfCPUProfile = flag.String("perf-cpu-profile", "", "Enable CPU profiling and write to specified file (e.g., cpu.prof)")
		perfMaxTraces = flag.Int("perf-max-traces", 1000, "Maximum number of operation traces to keep in memory")
		showVersion   = flag.Bool("version", false, "Show version and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("aztui version 0.0.1")
		return
	}

	// Initialize performance tracing if requested
	if *perfTrace {
		tracing.InitTracer(true, *perfMaxTraces)
		tracer := tracing.GetTracer()
		
		// Start CPU profiling if requested
		if *perfCPUProfile != "" {
			cpuProfilePath := *perfCPUProfile
			if err := tracer.StartCPUProfile(cpuProfilePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting CPU profile: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "CPU profiling enabled, writing to %s\n", cpuProfilePath)
		}
		
		// Set up signal handler to output stats on exit
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Fprintf(os.Stderr, "\n=== Performance Statistics ===\n")
			tracer.OutputStats()
			tracer.StopCPUProfile()
			os.Exit(0)
		}()
		
		fmt.Fprintf(os.Stderr, "Performance tracing enabled (max traces: %d)\n", *perfMaxTraces)
	} else {
		tracing.InitTracer(false, 1000)
	}

	a := NewAzTuiState()

	if err := a.AppLayout.App.SetRoot(a.AppLayout.Grid, true).Run(); err != nil {
		// Output performance stats before exiting on error
		if *perfTrace {
			tracer := tracing.GetTracer()
			fmt.Fprintf(os.Stderr, "\n=== Performance Statistics (Error Exit) ===\n")
			tracer.OutputStats()
			tracer.StopCPUProfile()
		}
		panic(err)
	}

	// Output performance stats on normal exit
	if *perfTrace {
		tracer := tracing.GetTracer()
		fmt.Fprintf(os.Stderr, "\n=== Performance Statistics ===\n")
		tracer.OutputStats()
		tracer.StopCPUProfile()
	}
}
