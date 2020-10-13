package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"jb-fls-check/fls-check/check"
)

var startDate, endDate string

func init() {
	rootCmd.AddCommand(periodReport)
	periodReport.Flags().StringVarP(&startDate, "startDate", "", "", "use this flag to set the start date of the period report")
	periodReport.Flags().StringVarP(&endDate, "endDate", "", "", "use this flag to set the end date of the period report")
	periodReport.Flags().String("token", "", "use this flag to set the API token you need for your request (default to set in config) the token has to be set with '' ")
	periodReport.MarkFlagRequired("startDate")
	periodReport.MarkFlagRequired("endDate")

}

var periodReport = &cobra.Command{
	Use:   "periodReport",
	Short: "Use this command to run the license report for a period of time you define",
	Long: `This command checks with a start Date and an end Date you define the license usage to a percentage value
This usage percentage value can you fls-check with a threshold you define as a percentage value too`,

	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("debug") {
			debug = true
		}
		startDate := cmd.Flag("startDate").Value.String()
		endDate := cmd.Flag("endDate").Value.String()
		token := cmd.Flag("token").Value.String()
		if token == "" {
			token = viper.GetString("token")
		}
		url := buildURL(https, hostname, endpoint, "hostname.report_endpoint", "/reportapi")
		errSlice, responseWeekly := check.GetWeeklyUsageReport(url, token, startDate, endDate, 90, debug)
		check.OutputMonitoring(errSlice, "weekly report checked", responseWeekly)

		err := cmd.Help()
		if err != nil {
			return
		}
	},
}
