`arduino-cli` supports command-line completion (also known as _tab completion_) for basic commands. Currently only
`bash`, `zsh`, `fish` shells are supported

### Before you start

In order to generate the file required to make the completion work you have to [install](installation.md) Arduino CLI
first.

### Generate the completion file

To generate the completion file you can use `arduino-cli completion [bash|zsh|fish] [--no-descriptions]`. By default
this command will print on the standard output (the shell window) the content of the completion file. To save to an
actual file use the `>` redirect symbol.

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
2. add `fpath=($HOME/completion_zsh $fpath)` at the beginning of your `~/.zshrc` file
3. `mv _arduino-cli ~/completion_zsh/`

Remember to open a new shell to test the functionality.

### Fish

Use `arduino-cli completion fish > arduino-cli.fish` to generate the completion file. At this point you can place the
file in `~/.config/fish/completions` as stated in the
[official documentation](http://fishshell.com/docs/current/index.html#where-to-put-completions). Remember to create the
directory if it's not already there `mkdir -p ~/.config/fish/completions/` and then place the completion file in there
with `mv arduino-cli.fish ~/.config/fish/completions/`

Remember to open a new shell to test the functionality.

#### Disabling command and flag descriptions

By default fish and zsh completion have command and flag description enabled by default. If you want to disable this
behaviour you can simply pass the `--no-descriptions` flag when calling `completion` command and the generated file will
not have descriptions

_N.B._ This flag is not compatible with bash

### Brew

If you install the `arduino-cli` using [homebrew](https://brew.sh/) package manager the completion should work out of
the box if you have followed the [official documentation](https://docs.brew.sh/Shell-Completion).
