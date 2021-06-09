package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/gnames/gndoc"
	"github.com/spf13/cobra"
)

func versionFlag(cmd *cobra.Command) bool {
	version, err := cmd.Flags().GetBool("version")
	if err != nil {
		log.Fatal(err)
	}
	if version {
		fmt.Printf("\nversion: %s\n\nbuild:   %s\n\n",
			gndoc.Version, gndoc.Build)
		return true
	}
	return false
}

func formatFlag(cmd *cobra.Command) {
	f, err := cmd.Flags().GetString("format")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if f != "" {
		opts = append(opts, gndoc.OptFormat(f))
	}
}

func tikaURLFlag(cmd *cobra.Command) {
	f, err := cmd.Flags().GetString("tika_url")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if f != "" {
		opts = append(opts, gndoc.OptTikaURL(f))
	}
}

func portFlag(cmd *cobra.Command) int {
	webPort, err := cmd.Flags().GetInt("port")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if webPort > 0 {
		opts = append(opts, gndoc.OptPort(webPort))
	}
	return webPort
}
