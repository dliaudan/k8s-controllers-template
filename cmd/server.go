package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

var serverPort int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server",
	Long: `Start a FastHTTP server that can handle HTTP requests.
This server is optimized for high-performance scenarios and can be used
for serving API endpoints or web content.`,
	Run: func(cmd *cobra.Command, args []string) {
		handler := func(ctx *fasthttp.RequestCtx) {
			switch string(ctx.Path()) {
			case "/":
				fmt.Fprintf(ctx, "Hello from FastHTTP! Server is running on port %d", serverPort)
			case "/health":
				ctx.SetStatusCode(fasthttp.StatusOK)
				fmt.Fprintf(ctx, "OK")
			case "/api/v1/status":
				ctx.SetContentType("application/json")
				fmt.Fprintf(ctx, `{"status":"running","port":%d}`, serverPort)
			default:
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				fmt.Fprintf(ctx, "Not Found")
			}
		}

		addr := fmt.Sprintf(":%d", serverPort)
		log.Info().Msgf("Starting FastHTTP server on %s", addr)
		log.Info().Msg("Available endpoints:")
		log.Info().Msg("  GET / - Main page")
		log.Info().Msg("  GET /health - Health check")
		log.Info().Msg("  GET /api/v1/status - API status")

		if err := fasthttp.ListenAndServe(addr, handler); err != nil {
			log.Error().Err(err).Msg("Error starting FastHTTP server")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on")
}
