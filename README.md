# UPSM

## Universal Playdate System Modification

![Black and White Pixel Art Version of the Playdate.](icon.png)

UPSM is a universal format to package system modifications that can be easily
installed by installers like the yapOS Installer.

## Spec

### Layout

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

### upsminfo Options

The `upsminfo` file is required to be in the `.upsm` directory even if no options
are specified.

#### loader?: bool

The `loader` option determines if the contained `Launcher.pdx` is a loader. If it
is a loader then all of the other modifications' `Launcher.pdx` folders will be
put into the `/System/Launchers/` folder instead of directly into `/System/`. If
the `loader` option is not specified then the default value is `false`.

#### name?: string

The `name` option will be copied into `/System/modinfo.json` to give info about the
installed modifications to any on-device updater or other tools.

#### version?: string

The `version` option will be copied into `/System/modinfo.json` to give info about
the installed modifications to any on-device updater or other tools.

#### id?: string

The `id` option should be a URL in reverse DNS notation and will be copied
into `/System/modinfo.json` to give info about the installed modifications to any
on-device updater or other tools.

#### author?: string

The `author` option will be copied into `/System/modinfo.json` to give info about
the installed modifications to any on-device updater or other tools.

#### description?: string

The `description` option will be copied into `/System/modinfo.json` to give info
about the installed modifications to any on-device updater or other tools.

#### launchers?: []string

The `launchers` option is an array of strings containing the paths of different launchers
that the modification contains.

### geninfo Options

The name of the `.gen` directory tells the installer what system app to patch. For
example if a `.gen` directory was named `ExampleApp.gen` then it would patch `/System/ExampleApp.pdx`.

#### patch?: string

`patch` is an optional option that is the path to the patch file that should be used
to patch part of PlaydateOS. If `patch` is not specified then the first `.patch`
file will be used.

#### options: string

The `options` option is simply the options passed to the `patch` that were used for
this modification.

#### path: string

The `outpath` option specifies where to copy the contents of the selected directory
to patch before patching.

## Distributing

When distributing a UPSM modification, you should distribute it as a `.upsm.zip`.
It can be either zipped with a single directory within it or with all the files from
the original `.upsm` in the root of the `.upsm.zip`
