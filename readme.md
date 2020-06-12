# peakbagger-tools

This project provides a cli to interact with the peakbagger.com website
It can:
 - Add 1 or several ascents to peakbagger from a Strava activity
 - Delete 1 or several ascents from peakbagger
 - List/export climber ascents

 The initial goal of this project was to provide a cli tool to automatically add ascents to peakbagger.com from Strava. The tool will:
  - Extract the GPX activity from Strava
  - Find the peaks crossed by the track
  - Register them on peakbagger

# How to build

- Provide your Strava client id and secret. Edit file `./pbtools/config/config.go`
```
...
type Config struct {
	HTTPPort int

	StravaClientID int    `config:"<your_client_id>,env=STRAVA_CLIENT_ID"`
	StravaSecretID string `config:"<your_secret_id>,env=STRAVA_SECRET_ID"`
...
```

- Build the tool
```
make build
```

# How to use

## Add ascents from a Strava activity
```
./bin/peakbagger add -activity https://www.strava.com/activities/<activityId>
```

## Remove an ascent
```
./bin/peakbagger delete -id <peakbaggger_aid>
```

## List ascents
```
./bin/peakbagger list -format csv -output my_ascents.csv
```



This is a very experimental project, and you might have issues running this tool.