package cmd

import (
	"log"

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
	importCmd.Flags().StringP("pagenum", "p", "", "Page number to import (default = all)")
	importCmd.Flags().StringP("calendarid", "c", "", "CalendarID of the google calendar to import data into.")

	if err := viper.BindPFlag("inputfile", importCmd.Flags().Lookup("inputfile")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	if err := viper.BindPFlag("employee", importCmd.Flags().Lookup("employee")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	if err := viper.BindPFlag("pagenum", importCmd.Flags().Lookup("pagenum")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	viper.SetDefault("pagenum", "-1")
	if err := viper.BindPFlag("calendarid", importCmd.Flags().Lookup("calendarid")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
}

func runImport(cmd *cobra.Command, args []string) {
	inputFile := viper.GetString("inputfile")
	employeeName := viper.GetString("employee")
	page := viper.GetInt("pagenum")
	// calendarId := viper.GetString("calendarid")

	log.Printf("Starting to parse file %s\n", inputFile)

	parser := parser.New(employeeName, inputFile)
	entries := parser.ProcessPages(page)
	log.Printf("%v", entries)

	// calendarClient = importer.NewCalendarClient(calendarId)

}
