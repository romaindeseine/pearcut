# Fly.io

On Fly.io, everything your app writes to stdout is automatically captured via [NATS](https://nats.io) and available via `fly logs`. To export these logs to an external service, Fly provides [fly-log-shipper](https://github.com/superfly/fly-log-shipper) — a pre-packaged app based on [Vector](https://vector.dev) under the hood.

```
Pearcut (stdout) → Fly.io (NATS) → fly-log-shipper (Vector) → destination
```

## Prerequisites

- Pearcut deployed on Fly.io with `--events=stdout`
- `flyctl` CLI authenticated

## Verify logs are flowing

```bash
fly logs --app pearcut
```

You should see JSON lines for each assignment event.

## Export with fly-log-shipper

Deploy the log shipper as a separate Fly app in your organization:

```bash
fly launch --image flyio/log-shipper:latest --no-public-ips
```

After generating `fly.toml`, update the internal port to `8686` (Vector's health check port).

Set the required secrets:

```bash
fly secrets set \
  ORG=your-org \
  ACCESS_TOKEN=$(fly tokens create readonly personal)
```

Then add the secrets for your chosen destination. See the [fly-log-shipper README](https://github.com/superfly/fly-log-shipper) for the full list of supported destinations and their required secrets.

The built-in destinations are mostly oriented towards monitoring and observability. For analytics destinations (BigQuery, ClickHouse, etc.), you will likely need a custom Vector configuration — see the [Vector ingestion guide](ingest-vector.md).

## Filter only Pearcut logs

By default, the log shipper captures logs from all apps in your org. To restrict to Pearcut, set the `SUBJECT` secret using the NATS subject pattern:

```bash
fly secrets set SUBJECT="logs.pearcut.>"
```

This captures all Pearcut logs across regions and instances.
