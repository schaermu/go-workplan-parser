package cmd

import (
	"log"

	"github.com/schaermu/workplan-parser/internal/importer"
	"github.com/schaermu/workplan-parser/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a workplan schedule into a google calendar.",
	Long: `You can import the data from a pdf file to your own google calendar.
	
Make sure you have a properly set up credentials.json file.`,
	Run: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("inputfile", "i", "", "Path to pdf file (can also be supplied to stdin).")
	importCmd.Flags().StringP("employee", "e", "", "Name of the employee to export data for")
	importCmd.Flags().Int32P("page", "p", -1, "Page number to import (default = all)")
	importCmd.Flags().StringP("calendarid", "c", "", "CalendarID of the google calendar to import data into.")
	importCmd.Flags().Bool("dry-run", false, "Don't import anything, only print actions to stdout.")

	if err := viper.BindPFlag("inputfile", importCmd.Flags().Lookup("inputfile")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	if err := viper.BindPFlag("employee", importCmd.Flags().Lookup("employee")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	if err := viper.BindPFlag("calendarid", importCmd.Flags().Lookup("calendarid")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	if err := viper.BindPFlag("page", importCmd.Flags().Lookup("page")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	if err := viper.BindPFlag("dry-run", importCmd.Flags().Lookup("dry-run")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}

	viper.SetDefault("dry-run", false)
}

func runImport(cmd *cobra.Command, args []string) {
	inputFile := viper.GetString("inputfile")
	employeeName := viper.GetString("employee")
	page := viper.GetInt32("page")
	calendarId := viper.GetString("calendarid")
	isDryRun := viper.GetBool("dry-run")

	log.Println("Starting import run with params:")
	log.Printf("  inputFile=%s\n", inputFile)
	log.Printf("  employeeName=%s\n", employeeName)
	log.Printf("  page=%d\n", page)
	log.Printf("  calendarId=%s\n", calendarId)
	log.Printf("  isDryRun=%t\n", isDryRun)

	parser := parser.New(employeeName, inputFile)
	entries := parser.ProcessPages(page)

	log.Printf("Starting to import events to calendar %s\n", calendarId)
	calendarClient := importer.NewCalendarClient(calendarId)
	for _, pageEntries := range entries {
		for _, entry := range pageEntries.Entries {
			log.Printf("  Creating %s for %s...", entry.Description, entry.GetWorktime())

			if !isDryRun {
				calendarClient.CreateEvent(entry.Start, entry.End, entry.Description, entry.IsAllDayEvent)
			}
		}
	}
}
