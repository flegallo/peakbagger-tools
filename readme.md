# peakbagger-tools

For now this project simply provides a tool to automatically register ascents to peakbagger.com from a Strava activity.

The tool extract the GPX activity from Strava, find the peaks crossed by the track, and register them on peakbagger.

# How to use

- Build the tool
```
make build
```

- Run
```
./bin/peakbagger-tools -pbUser <your_peakbagger_username> -pbPwd <your_peakbagger_username> -activity https://www.strava.com/activities/<activityId>
```


This is a very experimental project, and you might have issues running this tool.