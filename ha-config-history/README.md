# HA Config History

Git-like change history without git for Home Assistant

Create a timeline of changes to yaml configuration in Home Assistant with a diff viewer restore capability.

## Features

- Watches backed up files for changes, saving a new backup immediately
- Restores old versions from within Home Assistant (via Addon)
- Compares any two versions of a configuration file with a built in diff viewer
- Configurable backup options to track the usual Home Assistant configuration, or add any files you want

## Screenshots

![Homescreen](https://github.com/eddymoulton/ha-config-history/raw/main/ha-config-history/assets/home.png)

## Installation

### Home Assistant Add-on

1. Add this repository to your Home Assistant add-on store:

   [![Open your Home Assistant instance and show the add add-on repository dialog with a specific repository URL pre-filled.](https://my.home-assistant.io/badges/supervisor_add_addon_repository.svg)](https://my.home-assistant.io/redirect/supervisor_add_addon_repository/?repository_url=https%3A%2F%2Fgithub.com%2Feddymoulton%2Fha-config-history)

   Or manually:
   - Navigate to **Settings** → **Add-ons** → **Add-on Store** → **⋮** (three dots) → **Repositories**
   - Add: `https://github.com/eddymoulton/ha-config-history`

2. Find "HA Config History" in the add-on store and click **Install**

4. After installation, click **Start**

5. (Optional) Enable **Start on boot** and **Watchdog**

6. Access the UI through the **Web UI** button or the sidebar panel

### Standalone Docker Container

#### Using Docker CLI

```bash
docker run -d \
  --name ha-config-history \
  -p 40613:40613 \
  -v /path/to/your/homeassistant/config:/config \
  -v /path/to/backup/storage:/data \
  --restart unless-stopped \
  ghcr.io/eddymoulton/ha-config-history:latest
```

## Configuration

### Home Assistant Add-on Configuration

All configuration is done via the web app

### General Settings

<img width="721" height="521" alt="image" src="https://github.com/eddymoulton/ha-config-history/raw/main/ha-config-history/assets/general-settings.png" />

| Setting                             | Description                                                                                                                                                                                              |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Home Assistant Config Directory** | Must be mapped to the location where Home Assistant stores it's `configuration.yaml` file (amongst others)                                                                                               |
| **Backup Directory**                | Location that the backed up files get stored                                                                                                                                                             |
| **Server Port**                     | Web UI port                                                                                                                                                                                              |
| **Cron Schedule**                   | Optional schedule to run a full check, simlar to what is done on startup. This job will only take a backup if there is changed content. You can use this if you are having issue with the file watching. |
| **Default Max Backups**             | The default number of backups per configuration file that will be kept. This can be overridden per config                                                                                                |
| **Default Max Age**                 | The default number of days old that backup files can be kept. This can be overridden per config                                                                                                          |

### Config Backup Options

<img width="727" height="692" alt="image" src="https://github.com/eddymoulton/ha-config-history/raw/main/ha-config-history/assets/config-backup-options.png" />

| Setting         | Description                                                                          |
| --------------- | ------------------------------------------------------------------------------------ |
| **Name**        | Display name only                                                                    |
| **Path**        | The path to the file that should be backed up. See Backup Type for more information. |
| **Backup Type** | One of: `Multiple`, `Single`, or `Directory`. See details below.                     |
| **Max Backups** | The number of backups per configuration file that will be kept.                      |
| **Max Age**     | The number of days old that backup files can be kept.                                |

#### Backup Type Details

##### Multiple

Tracks multiple configurations inside a single file (eg. automations.yaml)

The path should be directly to a YAML file.

ID Node needs to be set to the YAML node that will be used to compare different configurations.
Friendly Name Node will be what is displayed in the UI.

##### Single

Tracks a single configuration for a file (eg. configuration.yaml)

The path should be directly to a YAML file.

##### Directory

Tracks all files within a directory as single configurations.

The path should be a directory path that contains one or more YAML files.


###### Additional Options
| Setting                   | Description                                                                                                                                      |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Include File Patterns** | (optional) Only include files matching one of the provided glob patterns. All files are included by default                                      |
| **Exclude File Patterns** | (optional) Exclude files matching one of the provided glob patterns. No files are included by default. You can exclude previously included files |

## Usage

### File cleanup

File cleanup occurs immediately after running a backup.

If a file is never updated, old versions will never be cleaned up.

## Contributing

Contributions welcome - create an issue and/or raise a PR

