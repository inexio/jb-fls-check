# JB-FLS-CHECK

## Description

Monitoring check plugin to check health, version, connection, period report of the [jetbrains floating license server](https://www.jetbrains.com/de-de/license-server/).  

## Features

GetVersionCheck

- sent a request against /check-version and checks if the server is running on the latest version
- you can set a bool value, if this value is set you will get a critical message if there is a new version available for the server. Otherwise you will get a warning message.

GetHealthCheck

- sent a request against /health and checks if the server health is okay

GetConnectionCheck

- sent a request against /check-connection and checks if the server can connect to all jetbrians services (like account service and homepage)

GetWeeklyUsageReport

- sent a request against /reportapi with a [api token you defined](https://www.jetbrains.com/help/license_server/detailed_server_usage_statistics.html#ad8) with the installation of the FLS.
- you have to sent this request with a start date and an end date, to define the period you will get the usage report.
- you have to define a usage threshold, that is the check value if this value is exceeded you wil get a warning message that say that the threshold for this toolpack is exceeded and you should move new licenses to the server

## Requirements

You need a floating license server from jetbrains running on your systems.

## Installation

```
go get github.com/inexio/jb-fls-check
```

or 

```
git clone git@github.com:inexio/jb-fls-check.git
```

or 

you download a [precompiled file](https://github.com/inexio/jb-fls-check/releases)

## Setup

After installation you have to setup your config or your environment variables.

### Configs

Default config file paths (3 paths): 

```
$HOME/.jb-fls-check
../config
/var/opt/jb-fls-check
```

You can set your token, hostname and api endpoints in a config and in environment variables. Otherwise you can set everything with flags on the CLI.
As you can see JB_FLS_CHECK is a prefix for the environment variable you can set and the hierachy like in a config you can define with an _ like (HOSTNAME_REPORT_ENDPOINT)

## Usage

### How to run jb-fls-check

First change directory to jb-fls-check

```
cd go/src/jb-fls-check
```

Then run main.go with one of  the following flags to call one check function

```
go run main.go --help
 	       --getHealth
	       --getConnection
	       --getVersion
```



For the period usage report use the following subcommand

```
go run main.go periodReport --startDate <yourStartDate> --endDate <yourEndDate> 
```

Use the --help flag to get all flags you can set with this subcommand

```
go run main.go periodReport --help
```
There is also a global --debug flag to debug the request you sent with the functions

## Getting Help

If there are any problems or something does not work as intended, open an issue on GitHub.

## Contribution

Contribution to the project are welcome.

We are looking forward to your bug reports, suggestions and fixes.

If you want to make any contributions make sure your go reports match up with our projects score **A+**.

When you contribute make sure you code is confirm to the **uber-go** coding style.

