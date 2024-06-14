# systemd

```
$> cp whosonfirst-spatial-pmtiles.service.example whosonfirst-spatial-pmtiles.service
```

Then adjust the settings in `whosonfirst-spatial-pmtiles.service` as necessary. Now install the service:

```
$> cp whosonfirst-spatial-pmtiles.service /etc/systemd/system/whosonfirst-spatial-pmtiles.service
$> systemctl daemon-reload
$> systemctl enable whosonfirst-spatial-pmtiles
$> systemctl start whosonfirst-spatial-pmtiles
```

To test, try something like:

```
$> curl -s http://localhost:9000/api/point-in-polygon -d '{"latitude": 37.621131, "longitude": -122.384292}' | jq '.places[]["wof:name"]'

"North America"
"United States of America"
"America/Los_Angeles"
"San Francisco International Airport"
"San Mateo"
"San Francisco-Oakland-San Jose"
"California"
"United States"
``
