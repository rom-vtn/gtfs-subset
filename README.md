# `gtfs-subset`

A simple CLI tool to take a localized subset of a larger GTFS feed which otherwise would be too big to parse without using a lot of memory.

## Syntax
Simply clone, `go build` and then use:
```bash
gtfs-subset \
    --minlat <MIN_LAT> \
    --maxlat <MAX_LAT> \
    --minlon <MIN_LON> \
    --maxlon <MAX_LON> \
    --input <path/to/input/feed.zip> \
    --output <path/to/output/feed.zip>
```

## How it works
- This file will take the subset of `stops.txt`, `trips.txt`, `routes.txt`, `stop_times.txt`, `calendar.txt`, `calendar_dates.txt`, `shapes.txt` and also includes some other files as-is (no need to have everything compressed)
- Note: the vast majority of the bloat comes from `stop_times.txt` and `shapes.txt` (if it exists), so once that's done the rest is pretty much optional
- Stop and trip IDs inside of the wanted area will be stored in-memory, so don't make it *too* large either (city/region wide is fine)

## Examples of large-ish feeds
> [!NOTE]
> None of the feeds below are mine, please check the licenses under which each producer releases their feeds before using them
- Feed for [all public transit in Germany](https://gtfs.de/de/feeds/de_nv/) from [gtfs.de](https://gtfs.de) (about 200MB)
- Feed for [all public the swiss railways](https://opendata.swiss/en/dataset/fahrplan-2025-gtfs2020) on [opendata.swiss](https://opendata.swiss) (about 120MB)
- The set of [all ÖBB transit](https://data.oebb.at/de/datensaetze~soll-fahrplan-gtfs~) (about 60 MB)

So far tested with the gtfs.de feed, taking a subset took about 1 minute.

## Spec-compliance
The following elements are taken into account, which makes for the core of the spec:
- `stops`
- `trips`
- `stop_times`
- `shapes`
- `routes`
- `agency`
- `feed_info`
- `attributions`

There might be some additions in the future, but this covers the core of the spec. Feel free to open an issue or send a PR if something's missing in a feed.

The validator at https://gtfs-validator.mobilitydata.org/ can be used to check the validity of a produced feed.