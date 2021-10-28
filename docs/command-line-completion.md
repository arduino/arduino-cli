`arduino-cli` supports command-line completion (also known as _tab completion_) for basic commands. Currently `bash`,
`zsh`, `fish`, and `powershell` shells are supported

### Before you start

In order to generate the file required to make the completion work you have to [install](installation.md) Arduino CLI
first.

### Generate the completion file

To generate the completion file you can use `arduino-cli completion [bash|zsh|fish|powershell] [--no-descriptions]`. By
default this command will print on the standard output (the shell window) the content of the completion file. To save to
an actual file use the `>` redirect symbol.

### Bash

Use `arduino-cli completion bash > arduino-cli.sh` to generate the completion file. At this point you can move that file
in `/etc/bash_completion.d/` (root access is required) with `sudo mv arduino-cli.sh /etc/bash_completion.d/`.

A not recommended alternative is to source the completion file in `~/.bashrc`.

Remember to open a new shell to test the functionality.

### Zsh

Use `arduino-cli completion zsh > _arduino-cli` to generate the completion file. At this point you can place the file in
a directory listed in your `fpath` if you have already created a directory to store your completion.

Or if you want you can create a directory, add it to your `fpath` and copy the file in it:

1. `mkdir ~/completion_zsh`
1. add `fpath=($HOME/completion_zsh $fpath)` at the beginning of your `~/.zshrc` file
1. `mv _arduino-cli ~/completion_zsh/`

Remember to open a new shell to test the functionality.

### Fish

Use `arduino-cli completion fish > arduino-cli.fish` to generate the completion file. At this point you can place the
file in `~/.config/fish/completions` as stated in the
[official documentation](http://fishshell.com/docs/current/index.html#where-to-put-completions). Remember to create the
directory if it's not already there `mkdir -p ~/.config/fish/completions/` and then place the completion file in there
with `mv arduino-cli.fish ~/.config/fish/completions/`

Remember to open a new shell to test the functionality.

### Powershell

Use `arduino-cli completion powershell > arduino-cli.ps1` to generate a temporary completion file. At this point you
need to add the content of the generated file to your PowerShell profile file.

1. `Get-Content -Path arduino-cli.ps1 | Add-Content -Path $profile` or add it by hand with your favourite text editor.
1. The previous command added two `using namespace` lines, move them on top of the `$profile` file.
1. If not already done, add the line `Set-PSReadlineKeyHandler -Key Tab -Function MenuComplete` to your `$profile` file:
   it is needed to enable the TAB completion in PowerShell.
1. `del arduino-cli.ps1` to remove the temporary file.

Remember to open a new shell to test the functionality.

For more information on tab-completion on PowerShell, please, refer to
[Autocomplete in PowerShell](https://techcommunity.microsoft.com/t5/itops-talk-blog/autocomplete-in-powershell/ba-p/2604524).

#### Disabling command and flag descriptions

By default fish, zsh and bash completion have command and flag description enabled by default. If you want to disable
this behaviour you can simply pass the `--no-descriptions` flag when calling `completion` command and the generated file
will not have descriptions

_N.B._ This flag is not compatible with powershell

### Brew

If you install the `arduino-cli` using [homebrew](https://brew.sh/) package manager the completion should work out of
the box if you have followed the [official documentation](https://docs.brew.sh/Shell-Completion).
