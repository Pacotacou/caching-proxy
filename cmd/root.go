package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"caching-proxy/proxy"

	"github.com/spf13/cobra"
)

var (
	port       int
	originURL  string
	clearCache bool
)

var rootCmd = &cobra.Command{
	Use:   "caching-proxy",
	Short: "A caching HTTP proxy server",
	Long: `A caching HTTP proxy server that forwards requests to a specified origin server
and caches the responses. If the same request is made again, it returns the cached
response instead of forwarding the request to the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create a new proxy server
		proxyServer := proxy.NewProxyServer(port, originURL)

		// If --clear-cache flag is set, clear the cache and exit
		if clearCache {
			if port == 0 {
				fmt.Println("Error: --port flag is required when using --clear-cache")
				cmd.Help()
				os.Exit(1)
			}

			clearURL := fmt.Sprintf("http://localhost:%d/admin/cache", port)
			req, err := http.NewRequest(http.MethodDelete, clearURL, nil)
			if err != nil {
				fmt.Printf("Error creating request: %v\n", err)
				os.Exit(1)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error clearing cache : %v\n", err)
				fmt.Println("Is the proxy running on the specified port?")
				os.Exit(1)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
			return
		}

		// Validate required flags
		if port == 0 {
			fmt.Println("Error: --port flag is required")
			cmd.Help()
			os.Exit(1)
		}

		if originURL == "" {
			fmt.Println("Error: --origin flag is required")
			cmd.Help()
			os.Exit(1)
		}

		// Start the proxy server
		if err := proxyServer.Start(); err != nil {
			fmt.Printf("Error starting proxy server: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVar(&port, "port", 0, "Port on which the proxy server will run")
	rootCmd.Flags().StringVar(&originURL, "origin", "", "URL of the server to which requests will be forwarded")
	rootCmd.Flags().BoolVar(&clearCache, "clear-cache", false, "Clear the cache and exit")
}
