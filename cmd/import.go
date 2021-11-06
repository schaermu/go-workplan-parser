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
	Long: `Allows you to parse and import a PDF file directly to Google calendar.

It is HIGHLY advised to run the import using --dry-run first to make sure the name is recognized and the fuzziness is configured correctly.
If you want to debug the output, check the folder temp/workplan-parser-[RANDOMID] to check on the detection boxes.
	
If you want to finally import to Google, make sure you have a properly set up credentials.json file within this folder containing your private key for Google API.`,
	Run: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("inputfile", "i", "", "Path to pdf file (can also be supplied to stdin).")
	importCmd.Flags().StringP("employee", "e", "", "Name of the employee to export data for")
	importCmd.Flags().Int32P("page", "p", -1, "Page number to import (default = all)")
	importCmd.Flags().Int32P("fuzziness", "f", 15, "Detection fuzziness, increase for sharper scans (default = 5, has to be odd)")
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
	if err := viper.BindPFlag("fuzziness", importCmd.Flags().Lookup("fuzziness")); err != nil {
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
	page := viper.GetInt("page")
	fuzziness := viper.GetInt("fuzziness")
	calendarID := viper.GetString("calendarid")
	isDryRun := viper.GetBool("dry-run")

	if fuzziness%2 == 0 {
		log.Fatalf("ERROR: invalid detection fuzziness %d, must be an odd number", fuzziness)
	}

	log.Println("\nStarting import run with params:")
	log.Printf("  inputFile=%s\n", inputFile)
	log.Printf("  employeeName=%s\n", employeeName)
	log.Printf("  page=%d\n", page)
	log.Printf("  fuzziness=%d\n", fuzziness)
	log.Printf("  calendarId=%s\n", calendarID)
	log.Printf("  isDryRun=%t\n\n", isDryRun)

	parser := parser.New(employeeName, inputFile, fuzziness)
	entries := parser.ProcessPages(page)

	log.Print("Starting to import events to calendar\n")
	calendarClient := importer.NewCalendarClient(calendarID)

	for _, pageEntries := range entries {
		if !isDryRun {
			calendarClient.ClearEvents(pageEntries.Month)
		}
		for _, entry := range pageEntries.Entries {
			log.Printf("  Creating %s for %s...", entry.Description, entry.GetWorktime())

			if !isDryRun {
				calendarClient.CreateEvent(entry.Start, entry.End, entry.Description, entry.IsAllDayEvent)
			}
		}
	}
}
