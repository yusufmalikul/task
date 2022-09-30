---
slug: /installation/
sidebar_position: 2
---

# Installation

Task offers many installation methods. Check out the available methods below.

## Package Managers

### Homebrew

If you're on macOS or Linux and have [Homebrew][homebrew] installed, getting
Task is as simple as running:

```bash
brew install go-task/tap/go-task
```

### Snap

Task is available in [Snapcraft][snapcraft], but keep in mind that your
Linux distribution should allow classic confinement for Snaps to Task work
right:

```bash
sudo snap install task --classic
```

### Chocolatey

If you're on Windows and have [Chocolatey][choco] installed, getting
Task is as simple as running:

```bash
choco install go-task
```

This installation method is community owned.


### Scoop

If you're on Windows and have [Scoop][scoop] installed, getting
Task is as simple as running:

```cmd
scoop install task
```

This installation method is community owned. After a new release of Task, it
may take some time until it's available on Scoop.

### AUR

If you're on Arch Linux you can install Task from
[AUR](https://aur.archlinux.org/packages/go-task-bin) using your favorite
package manager such as `yay`, `pacaur` or `yaourt`:

```cmd
yay -S go-task-bin
```

Alternatively, there's
[this package](https://aur.archlinux.org/packages/go-task) which installs from
the source code instead of downloading the binary from the
[releases page](https://github.com/go-task/task/releases):

```cmd
yay -S go-task
```

This installation method is community owned.

### Fedora

If you're on Fedora Linux you can install Task from the official
[Fedora](https://packages.fedoraproject.org/pkgs/golang-github-task/go-task/) repository using `dnf`:

```cmd
sudo dnf install go-task
```

This installation method is community owned. After a new release of Task, it
may take some time until it's available in [Fedora](https://packages.fedoraproject.org/pkgs/golang-github-task/go-task/).

### Nix

If you're on NixOS or have Nix installed
you can install Task from [nixpkgs](https://github.com/NixOS/nixpkgs):

```cmd
nix-env -iA nixpkgs.go-task
```

This installation method is community owned. After a new release of Task, it
may take some time until it's available in [nixpkgs](https://github.com/NixOS/nixpkgs).

### npm

You can also use Node and npm to install Task by installing
[this package](https://www.npmjs.com/package/@go-task/cli).

```bash
npm install -g @go-task/cli
```

## Get The Binary

### Binary

You can download the binary from the [releases page on GitHub][releases] and
add to your `$PATH`.

DEB and RPM packages are also available.

The `task_checksums.txt` file contains the SHA-256 checksum for each file.

### Install Script

We also have an [install script][installscript] which is very useful in
scenarios like CI. Many thanks to [GoDownloader][godownloader] for enabling the
easy generation of this script.

By default, it installs on the `./bin` directory relative to the working
directory:

```bash
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d
```

It is possible to override the installation directory with the `-b` parameter.
On Linux, common choices are `~/.local/bin` and `~/bin` to install for the
current user or `/usr/local/bin` to install for all users:

```bash
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
```

:::caution

On macOS and Windows, `~/.local/bin` and `~/bin` are not added to `$PATH` by
default.

:::

### GitHub Actions

If you want to install Task in GitHub Actions you can try using
[this action](https://github.com/arduino/setup-task)
by the Arduino team:

```yaml
- name: Install Task
  uses: arduino/setup-task@v1
```

This installation method is community owned.

## Build From Source

### Go Modules

First, make sure you have [Go][go] properly installed and setup.

You can easily install the latest release globally by running:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

Or you can install into another directory:

```bash
env GOBIN=/bin go install github.com/go-task/task/v3/cmd/task@latest
```

If using Go 1.15 or earlier, instead use:

```bash
env GO111MODULE=on go get -u github.com/go-task/task/v3/cmd/task@latest
```

:::tip

For CI environments we recommend using the [install script](#get-the-binary)
instead, which is faster and more stable, since it'll just download the latest
released binary.

:::

[go]: https://golang.org/
[snapcraft]: https://snapcraft.io/task
[homebrew]: https://brew.sh/
[installscript]: https://github.com/go-task/task/blob/master/install-task.sh
[releases]: https://github.com/go-task/task/releases
[godownloader]: https://github.com/goreleaser/godownloader
[choco]: https://chocolatey.org/
[scoop]: https://scoop.sh/
