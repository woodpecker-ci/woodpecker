# Migrate from Drone to Woodpecker

## Migrate from Drone >= v1.0.0

We currently do not provide a way to do so.
If you are interested or have a custom script to do so, please get in contact with us.

## Migrate from Drone <= v0.8

- Make sure you upgrade to Drone v0.8
- Upgrade to woodpecker v0.14.4, migration is run on startup
- If you use Sqlite3, rename `drone.sqlite` to `woodpecker.sqlite` and  
  Rename/Adjust volume-mount/folder from `/var/lib/drone/` to `/var/lib/woodpecker/`
- Upgrade to woodpecker v1.0.0, migration is run on startup
