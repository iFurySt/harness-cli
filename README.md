# harness-cli

`harness-cli` initializes a repository from the Harness project templates.

Run it in a new or existing repository:

```sh
harness-cli
```

The root command runs the same flow as:

```sh
harness-cli init
```

By default the CLI asks for a template language:

- `en`: `harness-template`
- `zh`: `harness-template-cn`

Useful non-interactive examples:

```sh
harness-cli init --language en
harness-cli init --language zh --target ./my-project
harness-cli init --language zh --force
harness-cli init --language en --dry-run
```

By default the CLI clones the selected template from GitHub. For local template development, point it at a local checkout explicitly:

```sh
harness-cli init --language zh --template-root ..
harness-cli init --language zh --source ../harness-template-cn
```

Existing files are left untouched unless their content is identical. Conflicting files stop the run; pass `--force` to overwrite them.

## Development

```sh
go test ./...
go run . --language en --dry-run
```
