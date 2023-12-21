# Local installation

You can install the Coder binary locally on any OS.

<div class="tabs">

## Linux

Install Coder on Linux or macOS using our
[install script](https://github.com/coder/coder/blob/main/install.sh):

```bash
curl -fsSL https://coder.com/install.sh | sh
```

You can preview what occurs during the install process:

```bash
curl -fsSL https://coder.com/install.sh | sh -s -- --dry-run
```

Modify the installation process by including flags. Run the help command for
reference:

```bash
curl -fsSL https://coder.com/install.sh | sh -s -- --help
```

## macOS

Install Coder on macOS from our official
[Homebrew tap](https://github.com/coder/homebrew-coder):

```bash
brew install coder/coder/coder
```

## Windows

Install Coder on Windows using the official installer:

1. Download the Windows installer from
   [GitHub releases](https://github.com/coder/coder/releases/latest) or from
   `winget`

   ```powershell
    winget install Coder.Coder
   ```

2. Run the application

![Windows installer](../images/install/windows-installer.png)

</div>

### Reordering your PATH

Here's the path configuration for users installing on macOS or Linux.

<div class="tabs">

## bash

Configure your path in `~/.bashrc`

```bash
export PATH="/opt/homebrew/bin:$PATH"
```

ℹ If you ran install.sh with a `--prefix` flag, you can replace `/opt/homebrew`
with whatever value you used there. Make sure to leave the `/bin` at the end!

## fish

Configure your coder path in `~/.config/fish/config.fish`

```shell
fish_add_path "/opt/homebrew/bin"
```

ℹ If you ran install.sh with a `--prefix` flag, you can replace `/opt/homebrew`
with whatever value you used there. Make sure to leave the `/bin` at the end!

## zsh

Configure your path in `~/.zshrc`

```shell
export PATH="/opt/homebrew/bin:$PATH"
```

ℹ If you ran install.sh with a `--prefix` flag, you can replace `/opt/homebrew`
with whatever value you used there. Make sure to leave the `/bin` at the end!

</div>

You can observe that the order has changed:

```console
which -a coder
/opt/homebrew/bin/coder
/usr/local/bin/coder
```

## Start a Coder server

The [`coder server`](../cli/server.md) command starts your server and opens an
external access URL.

```bash
# Automatically sets up an external access URL on *.try.coder.app
coder server

# Requires a PostgreSQL instance (version 13 or higher) and external access URL
coder server --postgres-url <url> --access-url <url>
```

> Set `CODER_ACCESS_URL` to the external URL that users and workspaces will use
> to connect to Coder. This is not required if you are using the tunnel. Learn
> more about Coder's [configuration options](../admin/configure.md).

### Uninstall coder

If you want to uninstall a version of `coder` that you installed with a package
manager, you can run whichever one of these commands applies:

<div class="tabs">

## macOS

Remove Coder using [`brew`](https://brew.sh/):

```shell
brew uninstall coder
```

## Debian/Ubuntu

Remove Coder using `dpkg`:

```shell
sudo dpkg -r coder
```

## Fedora

Remove Coder using `rpm`:

```shell
sudo rpm -e coder
```

## Alpine

Remove Coder using `apk`:

```shell
sudo apk del coder
```

</div>

## Next steps

- [Configuring Coder](../admin/configure.md)
- [Templates](../templates/index.md)
