# Miry

Miry can retrieve the latest mirror list from [MirrorStatus](https://www.archlinux.org/mirrors/status/) page, filter out inactive and non HTTP mirrors, sort them by speed and write output to file.
The idea is ~~stolen~~ heavily inspired by [Reflector](https://wiki.archlinux.org/index.php/Reflector) with some adjustments - it's written in **Go**, allows you to rate multiple mirrors concurrently and has a progress bar!

##### Progress bar example

```
██████████████████████████████████████████░░░░░░░░░░░ 78% | ETA: 4s | 78/100
```

## CLI options

| `Flag`                | Default value              | Description |
|-----------------------|----------------------------|-------------|
| `--jobs`              | `4`                        | Specify number of concurrently running rating jobs. By increasing it you gonna decrease program execution time and rating accuracy |
| `--limit`             | `-1`                       | Limit number of mirrors to rate, -1 means no limit |
| `--timeout`           | `5000`                     | Specify number of milliseconds to wait before cancelling mirror rating |
| `--progress`          | `true`                     | Enable rating progress bar |
| `--backup`            | `true`                     | Enable output file backup |
| `--output`            | `/etc/pacman.d/mirrorlist` | Specify output file (supports relative paths) |
| `--help` (alias `-h`) | `5000`                     | Print help |
