package check

import (
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-resty/resty/v2"
	"github.com/inexio/go-monitoringplugin"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
HealthResponse is a struct for parsing the json response of the /health request
*/
type HealthResponse struct {
	ServerUID    string `json:"serverUID"`
	LastCallHome string `json:"lastCallHome"`
}

/*
VersionResponse is a struct for parsing the json response of the /check-version request
*/
type VersionResponse struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	AvailableUpdate bool   `json:"updateAvailable"`
}

/*
OverallReport  is a struct for parsing the json response of the /reportapi request
*/
type OverallReport struct {
	Report []WeeklyReport `json:"Overall"`
}

/*
WeeklyReport  is a struct for parsing the json hierachy of the /reportapi request
*/
type WeeklyReport struct {
	License      string `json:"License"`
	MaxUsage     int    `json:"Max usage"`
	MaxAvailable int    `json:"Max available"`
}

/*
ErrorAndCode is a struct to create a an array with error and an monitoring exit code
*/
type ErrorAndCode struct {
	ExitCode int
	Error    error
}

/*
GetHealthCheck is a function to sent a request against fls api and parse and check the /health response
*/
func GetHealthCheck(url string, debug bool) []ErrorAndCode {
	var errSlice []ErrorAndCode
	if url == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("The url must not be empty")})
		return errSlice
	}
	c := resty.New()
	request := c.SetDebug(debug).SetDebugBodyLimit(1000).R()
	response, err := request.Get(url)
	if err != nil {
		errSlice = append(errSlice, ErrorAndCode{3, errors.Wrap(err, "error during http request")})
		return errSlice
	}
	var resp HealthResponse
	err = json.Unmarshal(response.Body(), &resp)

	if err != nil {
		errSlice = append(errSlice, ErrorAndCode{3, errors.Wrap(err, response.Status())})
		return errSlice
	}
	if resp.ServerUID == "" || resp.LastCallHome == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("Server: " + resp.ServerUID + "did call home last time: " + resp.LastCallHome)})
		return errSlice
	}
	errSlice = append(errSlice, ErrorAndCode{0, errors.New("Successfully connected to server: " + resp.ServerUID + " its last call home was " + resp.LastCallHome)})
	return errSlice
}

/*
GetConnectionCheck is a function to sent a request against fls api and parse and check the /check-connection response
*/
func GetConnectionCheck(url string, debug bool) []ErrorAndCode {
	var errSlice []ErrorAndCode
	if url == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("This URL must be not empty")})
		return errSlice
	}
	c := resty.New()
	request := c.SetDebug(debug).SetDebugBodyLimit(1000).R()
	response, err := request.Get(url)
	if err != nil {
		errSlice = append(errSlice, ErrorAndCode{3, errors.Wrap(err, "error during http request")})
		return errSlice
	}
	accountConnection, _ := regexp.Match("https://account.jetbrains.com	OK", response.Body())
	websiteConnection, _ := regexp.Match("https://www.jetbrains.com	OK", response.Body())
	if accountConnection != true || websiteConnection != true {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("No connection to service if server is running")})
		return errSlice
	}
	errSlice = append(errSlice, ErrorAndCode{0, errors.New("connection to account services and to the homepage is successful")})
	return errSlice

}

/*
GetVersionCheck is a function to sent a request against fls api and parse and check the /check-version response
*/
func GetVersionCheck(url string, throwCritical bool, debug bool) []ErrorAndCode {
	var errSlice []ErrorAndCode
	if url == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("This URL must be not empty")})
		return errSlice
	}
	c := resty.New()
	request := c.SetDebug(debug).SetDebugBodyLimit(1000).R()
	response, err := request.Get(url)
	if err != nil {
		errSlice = append(errSlice, ErrorAndCode{3, errors.Wrap(err, "error during http request")})
		return errSlice
	}
	var resp VersionResponse
	err = json.Unmarshal(response.Body(), &resp)

	if err != nil {
		errSlice = append(errSlice, ErrorAndCode{2, errors.Wrap(err, response.Status())})
		return errSlice
	}

	if resp.LatestVersion == "" || resp.CurrentVersion == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("Could not parse response into response struct")})
		return errSlice
	} else if resp.AvailableUpdate == true && resp.CurrentVersion != resp.LatestVersion {
		if throwCritical != true {
			errSlice = append(errSlice, ErrorAndCode{1, errors.New("Server is running on this version: " + resp.CurrentVersion + " this version is the latest version: " + resp.LatestVersion)})
			return errSlice
		}
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("Server is running on this version" + resp.CurrentVersion + " install this version as soon as possible!: " + resp.LatestVersion)})
		return errSlice

	} else {
		errSlice = append(errSlice, ErrorAndCode{0, errors.New("Server is running on version: " + resp.CurrentVersion)})
		return errSlice
	}
}

/*
GetWeeklyUsageReport is a function to sent a request against fls api and parse and check the /reportapi response with a calculation of the usage percentage to your threshold if exceeded you will get a warning message
*/
func GetWeeklyUsageReport(url string, token string, startDate string, endDate string, threshold int, debug bool) ([]ErrorAndCode, []monitoringplugin.PerformanceDataPoint) {
	var errSlice []ErrorAndCode
	var resp OverallReport
	if url == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("This URL must not be empty")})
	}
	if token == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("This token must not be empty")})
	}
	if startDate == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("This start date must not be empty (YYYY-MM-DD)")})
	} else {
		_, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			errSlice = append(errSlice, ErrorAndCode{2, errors.Wrap(err, "The start date is not in the right syntax use (YYYY-MM-DD)")})
		}
	}
	if endDate == "" {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("This end date must not be empty (YYYY-MM-DD)")})
	} else {
		_, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			errSlice = append(errSlice, ErrorAndCode{2, errors.Wrap(err, "The end date is not in the right syntax use (YYYY-MM-DD)")})
		}
	}
	if threshold <= 0 || threshold >= 100 {
		errSlice = append(errSlice, ErrorAndCode{2, errors.New("The threshold have to be greater than 0 and lower than 100")})

	}
	if len(errSlice) > 0 {
		return errSlice, nil
	}
	c := resty.New()
	request := c.SetDebug(debug).SetDebugBodyLimit(1000).R()
	response, err := request.SetQueryParams(map[string]string{
		"granularity": "0",
		"start":       startDate,
		"end":         endDate,
		"token":       token,
	}).Post(url)

	if err != nil {
		errSlice = append(errSlice, ErrorAndCode{3, errors.Wrap(err, "error during http request")})
		return errSlice, nil
	}
	err = json.Unmarshal(response.Body(), &resp)

	if err != nil {
		errSlice = append(errSlice, ErrorAndCode{2, errors.Wrap(err, response.Status())})
		return errSlice, nil
	}
	var performanceDataSlice []monitoringplugin.PerformanceDataPoint
	for i := 0; i < len(resp.Report); i++ {
		r := regexp.MustCompile("[^a-z]")
		licenseString := r.ReplaceAllString(strings.ToLower(resp.Report[i].License), "_")
		maxUsage := "max_usage_" + licenseString
		maxAvailable := "max_available_" + licenseString
		performanceDataSlice = append(performanceDataSlice, *monitoringplugin.NewPerformanceDataPoint(maxUsage, float64(resp.Report[i].MaxUsage), ""))
		performanceDataSlice = append(performanceDataSlice, *monitoringplugin.NewPerformanceDataPoint(maxAvailable, float64(resp.Report[i].MaxAvailable), ""))

		if resp.Report[i].MaxAvailable == 0 {
			if resp.Report[i].MaxUsage > 0 {
				errSlice = append(errSlice, ErrorAndCode{2, errors.New("Usage for " + resp.Report[i].License + "must be greater than 0 if there are no licenses on the server")})
				continue
			} else {
				continue
			}
		} else if resp.Report[i].MaxAvailable > 0 {
			percentageValue := (resp.Report[i].MaxUsage / resp.Report[i].MaxAvailable) * 100
			if percentageValue >= threshold {
				errSlice = append(errSlice, ErrorAndCode{1, errors.New("Your threshold for " + resp.Report[i].License + "is exceeded, please fls-check the licenses on the server")})
			} else {
				errSlice = append(errSlice, ErrorAndCode{0, errors.New("The Licenses Usage for " + resp.Report[i].License + " is " + strconv.Itoa(percentageValue) + "%")})
			}
		}
	}
	return errSlice, performanceDataSlice
}

/*
OutputMonitoring is a function to handle the lib go-monitoringplugin and output the ErrorAndCode struct and the performance data
*/
func OutputMonitoring(errSlice []ErrorAndCode, defaultMessage string, performanceDataSlice []monitoringplugin.PerformanceDataPoint) {
	response := monitoringplugin.NewResponse(defaultMessage)
	for i := 0; i < len(errSlice); i++ {
		response.UpdateStatus(errSlice[i].ExitCode, errSlice[i].Error.Error())
	}
	for i := 0; i < len(performanceDataSlice); i++ {
		err := response.AddPerformanceDataPoint(&performanceDataSlice[i])
		if err != nil {
			spew.Dump(err)
		}
	}
	response.OutputAndExit()
}
