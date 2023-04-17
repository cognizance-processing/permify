package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"permify/internal"
)

// NewVersionCommand - Creates new Version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "prints the permify version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s\n", internal.Version)
			return nil
		},
	}
}
