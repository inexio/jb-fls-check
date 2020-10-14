package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"jb-fls-check/fls-check/check"
	"os"
	"strings"
)

var (
	cfgFile, hostname, endpoint string
	https, debug                bool
)

var rootCmd = &cobra.Command{
	Use:   "Check Functions",
	Short: "Checks License Usage and if server is reachable",
	Long: `Checks if license server can call home to JetBrains.
And if the server is connected to the account services from JetBrains, the current license usage and the max availability of licenses on the server`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("https") {
			https = true
		}

		if viper.GetBool("debug") {
			debug = true
		}

		check.Client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: viper.GetBool("insecure-ssl-cert")})

		if viper.GetBool("getHealth") {
			url := buildURL(https, hostname, endpoint, "hostname.health_endpoint", "/health")
			errSlice := check.GetHealthCheck(url, debug)
			check.OutputMonitoring(errSlice, "server health checked", nil)
		}
		if viper.GetBool("getConnection") {
			url := buildURL(https, hostname, endpoint, "hostname.connection_endpoint", "/check-connection")

			errSlice := check.GetConnectionCheck(url, debug)
			check.OutputMonitoring(errSlice, "server connection checked", nil)
		}
		if viper.GetBool("getVersion") {
			var throwCritical bool

			url := buildURL(https, hostname, endpoint, "hostname.version_endpoint", "/check-version")
			if viper.GetBool("throwCritical") {
				throwCritical = true
			}
			errSlice := check.GetVersionCheck(url, throwCritical, debug)
			check.OutputMonitoring(errSlice, "server version checked", nil)
		}
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", "", "use this flag to set hostname")
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "use this flag to set endpoint")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "use this flag to switch to debug mode")
	rootCmd.PersistentFlags().BoolVar(&https, "https", false, "use this flag set https prefix")
	rootCmd.PersistentFlags().Bool("insecure-ssl-cert", false, "accept an insecure ssl certificate")
	rootCmd.Flags().Bool("getHealth", false, "use this flag to get the health of the server")
	rootCmd.Flags().Bool("getConnection", false, "use this flag to fls-check if the server got connection to JetBrains Services")
	rootCmd.Flags().Bool("getVersion", false, "use this flag to fls-check if the server runs on the latest version")
	rootCmd.Flags().Bool("throwCritical", false, "use this flag to warn if there is a update available (if set there will be a critical warning to update the server, if not then just a warning message)")
	viper.BindPFlag("getReport", rootCmd.Flags().Lookup("getReport"))
	viper.BindPFlag("getHealth", rootCmd.Flags().Lookup("getHealth"))
	viper.BindPFlag("getConnection", rootCmd.Flags().Lookup("getConnection"))
	viper.BindPFlag("getVersion", rootCmd.Flags().Lookup("getVersion"))
	viper.BindPFlag("throwCritical", rootCmd.Flags().Lookup("throwCritical"))
	viper.BindPFlag("insecure-ssl-cert", rootCmd.PersistentFlags().Lookup("insecure-ssl-cert"))
}

func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(os.ExpandEnv("$HOME/go/src/jb-fls-check/fls-check/config"))
		viper.AddConfigPath("../config")
		viper.AddConfigPath("/var/opt/jb-fls-fls-check")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		replacer := strings.NewReplacer(".", "_")
		viper.SetEnvKeyReplacer(replacer)

		viper.SetEnvPrefix("JB_FLS_CHECK")
		viper.AutomaticEnv() // read in environment variables that match

		_ = viper.ReadInConfig()

	}
}

/*
Execute executes the command handler
*/
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func buildURL(https bool, hostname string, endpoint string, EndpointConfigString string, defaultEndpoint string) string {
	var prefix string
	var currentEndpoint string
	if endpoint != "" {
		currentEndpoint = endpoint
	}
	if currentEndpoint == "" && EndpointConfigString != "" {
		currentEndpoint = viper.GetString(EndpointConfigString)
	}
	if currentEndpoint == "" && defaultEndpoint != "" {
		currentEndpoint = defaultEndpoint
	}

	if https != true {
		prefix = "http://"
	} else {
		prefix = "https://"
	}
	if hostname == "" {
		hostname = viper.GetString("hostname.hostname")
	}

	url := prefix + hostname + currentEndpoint

	return url
}
