# Migrate from Drone to Woodpecker

## Migrate from Drone >= v1.0.0

We currently do not provide a way to do so.
If you are interested or have a custom script to do so, please get in contact with us.

## Migrate from Drone <= v0.8

- Make sure you are already running Drone v0.8
- Upgrade to Woodpecker v0.14.4, migration will be done during startup
- If you are using Sqlite3, rename `drone.sqlite` to `woodpecker.sqlite` and
  rename or adjust the mount/folder of the volume from `/var/lib/drone/`
  to `/var/lib/woodpecker/`
- Upgrade to Woodpecker v1.0.0, the migration will be performed during
  startup
