package cmd

import (
	"fmt"

	"github.com/schaermu/workplan-parser/internal/parser"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parses an Inselgruppe workplan.",
	Long:  `Parse a workplan from a pdf file and display the output.`,
	Run:   run,
}

var (
	inputFile    string
	employeeName string
	prettyPrint  bool
)

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "Path to pdf file (can also be supplied to stdin).")
	parseCmd.Flags().StringVarP(&employeeName, "employee-name", "e", "", "Name of the employee to export data for")
	parseCmd.Flags().BoolVarP(&prettyPrint, "pretty", "p", false, "Pretty-print parsed input.")
	parseCmd.Flags().BoolP("json", "j", true, "Print parsed input as json.")
}

func run(cmd *cobra.Command, args []string) {
	fmt.Printf("Starting to parse file %s\n", inputFile)

	parser := parser.New(employeeName, inputFile)
	parser.ProcessPages()
}
