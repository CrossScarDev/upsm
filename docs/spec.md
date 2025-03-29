# Specification

## Layout

UPSM is really just a directory similar to the macOS `.App` format and the Playdate
`.pdx` format. At the root of a UPSM folder is multiple `.pdx` folders, multiple
`.gen` folders, and a `upsminfo` file. All of the `.pdx` folders in the root folder
will be copied over to the `/System/` folder in the `.pdos` file unless a loader
is being installed along with the mod. If a loader is being installed along with
this mod then the `Launcher.pdx` folder will be copied over to `/System/Launchers/SOMENAME.pdx`.
The `upsminfo` file is formatted similar to a `pdxinfo` file and contains metadata
about the modification. The `.gen` folders contain parts of the modification that
patch built-in parts of PlaydateOS. These `.gen` folders contain a `.patch` file
and a `geninfo` file containing metadata.

## upsminfo Options

The `upsminfo` file is required to be in the `.upsm` directory even if no options
are specified.

### loader?: bool

The `loader` option determines if the contained `Launcher.pdx` is a loader. If it
is a loader then all of the other modifications' `Launcher.pdx` folders will be
put into the `/System/Launchers/` folder instead of directly into `/System/`. If
the `loader` option is not specified then the default value is `false`.

### name?: string

The `name` option will be copied into `/System/upsm_mods.json` to give info about
the installed modifications to any on-device updater or other tools.

### version?: string

The `version` option will be copied into `/System/upsm_mods.json` to give info about
the installed modifications to any on-device updater or other tools.

### url?: string

The `url` option will be copied into `/System/upsm_mods.json` to give info about
the installed modifications to any on-device updater or other tools.

### author?: string

The `author` option will be copied into `/System/upsm_mods.json` and is a url that
can be used to download updates from.

### description?: string

The `description` option will be copied into `/System/upsm_mods.json` to give info
about the installed modifications to any on-device updater or other tools.

### launcherpath?: string

The `launcherpath` option is used to specify what path the `Launcher.pdx` file should
use if there is a loader being used. It is only required when a `Launcher.pdx` folder
is present.

## geninfo Options

The name of the `.gen` directory tells the installer what system app to patch. For
example if a `.gen` directory was named `ExampleApp.gen` then it would patch `/System/ExampleApp.pdx`.

### options: string

The `options` option is simply the options passed to the `patch` that were used for
this modification.

### path: string

The `path` option specifies where to copy the contents of the selected directory
to patch before patching.

## Examples

### upsminfo

```txt
name=Example UPSM
version=0.2.1-example1
author=grady.link
description=This is a very cool example.
launcherpath=Cool-Launcher.pdx
```

### geninfo

```txt
options=-s -p1 -d src/
path=src/
```
