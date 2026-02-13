package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"local-dev-tool/internal/rafiki"
	"local-dev-tool/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func NewRafikiCmd() *cobra.Command {
	var skipUI bool
	var selectedAssets []string
	var waitForReady int

	cmd := &cobra.Command{
		Use:   "rafiki",
		Short: "Setup and seed Rafiki with assets and liquidity",
		Long:  `Setup and seed Rafiki backend with currency assets and initial liquidity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := rafiki.LoadConfig()

			if waitForReady > 0 {
				healthURL := strings.TrimSuffix(cfg.GraphQLEndpoint, "/graphql") + "/healthz"
				fmt.Printf("Waiting up to %ds for Rafiki to be ready at %s...\n", waitForReady, healthURL)

				client := &http.Client{Timeout: 2 * time.Second}
				deadline := time.Now().Add(time.Duration(waitForReady) * time.Second)

				for time.Now().Before(deadline) {
					resp, err := client.Get(healthURL)
					if err == nil {
						resp.Body.Close()
						if resp.StatusCode == http.StatusOK {
							fmt.Println("Rafiki is ready")
							break
						}
					}
					time.Sleep(2 * time.Second)
				}

				if time.Now().After(deadline) {
					return fmt.Errorf("rafiki did not become ready within %ds", waitForReady)
				}
			}

			var assetsToCreate []rafiki.Asset
			if len(selectedAssets) > 0 {
				// Use CLI-specified assets
				assetsToCreate = filterAssetsByCode(selectedAssets)
			} else if skipUI {
				// Use all assets if --skip-ui is set
				assetsToCreate = rafiki.DefaultAssets
			} else {
				// Show interactive TUI for asset selection
				assetSelector := tui.NewAssetSelectorModel(rafiki.DefaultAssets)
				p := tea.NewProgram(assetSelector)
				finalModel, err := p.Run()
				if err != nil {
					return fmt.Errorf("failed to run TUI: %w", err)
				}

				selected := finalModel.(tui.AssetSelectorModel)
				if selected.Cancelled {
					fmt.Println("Operation cancelled")
					return nil
				}

				assetsToCreate = selected.SelectedAssets()
			}

			if len(assetsToCreate) == 0 {
				fmt.Println("No assets selected")
				return nil
			}

			fmt.Printf("Setting up Rafiki at %s\n", cfg.GraphQLEndpoint)
			fmt.Printf("Creating %d assets...\n", len(assetsToCreate))

			// Create assets
			if err := rafiki.EnsureAssets(cfg, assetsToCreate); err != nil {
				return fmt.Errorf("failed to create assets: %w", err)
			}

			// Add liquidity
			if err := rafiki.EnsureLiquidity(cfg, assetsToCreate); err != nil {
				return fmt.Errorf("failed to add liquidity: %w", err)
			}

			fmt.Println("✅ Rafiki configuration complete")
			return nil
		},
	}

	cmd.Flags().BoolVar(&skipUI, "skip-ui", false, "Skip interactive UI and create all assets")
	cmd.Flags().StringSliceVar(&selectedAssets, "assets", []string{}, "Comma-separated list of asset codes (e.g., USD,EUR,GBP)")
	cmd.Flags().IntVar(&waitForReady, "wait-for-ready", 0, "Wait up to N seconds for Rafiki to be healthy before proceeding")

	return cmd
}

func filterAssetsByCode(codes []string) []rafiki.Asset {
	codeMap := make(map[string]bool)
	for _, code := range codes {
		codeMap[code] = true
	}

	var filtered []rafiki.Asset
	for _, asset := range rafiki.DefaultAssets {
		if codeMap[asset.Code] {
			filtered = append(filtered, asset)
		}
	}
	return filtered
}
