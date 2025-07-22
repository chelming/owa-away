# OWA Away Calendar Filter

This Go application fetches a calendar from a specified URL, processes it to remove events with specific summaries (Away, Tentative, and Free), and serves the modified calendar to be ingested by another calendaring tool. The application is designed to run in a Docker container.


## Environment Variables

### Required

- `URL`: The URL of the calendar to fetch and process.


### Optional

- `DISPLAY_NAME`: The display name for the calendar. Defaults to "My Calendar" if not set.
- `EVENT_TYPES`: Comma-separated list of event types to remove by summary (e.g., `Away,Tentative,Free`). If set to an empty string, no events will be deleted by summary. If unset, the default is to remove events with summaries containing "Away", "Tentative", or "Free".
- `DELETE_REGEXES`: One or more regular expressions (comma-separated) to match and remove calendar events. See below for usage and examples.

### DELETE_REGEXES Details & Examples

You can use the `DELETE_REGEXES` environment variable to specify one or more regular expressions (comma-separated) to match and remove calendar events. Each regex is applied to the raw event block (after line breaks are replaced with `%%`).

#### Examples

- Remove all events that start at exactly 00:01 or 00:31 (i.e., times ending in :01 or :31):

  ```
  DELETE_REGEXES='DTSTART(?:;TZID=[^:]+)?:\d{8}T\d{2}(?:3|0)100%%'
  ```

- Remove all events that start at midnight (00:00):

  ```
  DELETE_REGEXES='DTSTART(?:;TZID=[^:]+)?:\d{8}T\d{2}0000%%'
  ```

- Remove all events that start at 30 minutes past the hour (e.g., 00:30, 01:30, etc.):

  ```
  DELETE_REGEXES='DTSTART(?:;TZID=[^:]+)?:\d{8}T\d{2}3000%%'
  ```

- Remove all events that start during business hours (08:00–17:59):

  ```
  DELETE_REGEXES='DTSTART(?:;TZID=[^:]+)?:\d{8}T0[8-9]\d{4}%%,DTSTART(?:;TZID=[^:]+)?:\d{8}T1[0-7]\d{4}%%'
  ```

- Remove all events with a specific duration (e.g., 30 minutes):

  ```
  DELETE_REGEXES='DURATION:PT30M'
  ```

If any regex is invalid, it will be ignored and an error will be logged.
# OWA Away Calendar Filter

This Go application fetches a calendar from a specified URL, processes it to remove events with specific summaries (Away, Tentative, and Free), and serves the modified calendar to be ingested by another calendaring tool. The application is designed to run in a Docker container.


## Environment Variables

### Required

- `URL`: The URL of the calendar to fetch and process.

### Optional

- `DISPLAY_NAME`: The display name for the calendar. Defaults to "My Calendar" if not set.

## Running the Docker Container

To run the Docker container, use the following command:

```sh
docker run -e URL="https://example.com/calendar.ics" -e DISPLAY_NAME="My Custom Calendar" -p 8080:8080 ghcr.io/chelming/owa-away
```

Here is a sample `docker-compose.yml` file:

```yaml
services:
  owa-away:
    image: ghcr.io/chelming/owa-away
    environment:
      - URL=https://example.com/calendar.ics
      - DISPLAY_NAME=My Custom Calendar
      # Optional environment variables:
      - EVENT_TYPES=Away,Tentative,Free  # optional: comma-separated event types to remove by summary
      - DELETE_REGEXES=                  # optional: regex patterns to remove events
    ports:
      - "8080:8080"
```

## Application Overview

The application performs the following steps:

1. Fetches the calendar from the specified URL.
2. Processes the calendar to remove events with summaries containing "Away", "Tentative", or "Free".
3. Renames the calendar using the `DISPLAY_NAME` environment variable.
4. Serves the modified calendar on port 8080.

## License

This project is licensed under the MIT License.
