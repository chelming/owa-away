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
