# Project Rules

## Before making code changes

- Discuss the design with the user
- Show example changes to make sure we are in the right direction
- If the user request is not clear, ask for more details
- List possible alternatives to make the change

## Making code changes

- Do the minimal change needed, no extra code that is not needed right now
- Avoid unrelated changes (spelling, whitespace, etc.)
- Keep changes small to make human review easy and avoid mistakes

## After making code changes

- Run `make fmt` to format code with golangci-lint formatters
- Run `make test` to verify all tests pass
- Running specific package tests (e.g., `go test ./pkg/test/...`) is fine for quick local verification

## File organization

- Keep files focused - separate files for different concerns (e.g., `html.go`, `yaml.go`, `summary.go`)
- All files need SPDX license headers - check existing files for the format

## Error handling

- Check existing code for error formatting conventions

## Testing

- Use `helpers.FakeTime(t)` for time-dependent tests to ensure reproducibility

## Commit messages

When the user wants to commit changes, suggest a commit message.

The main purpose is to explain why the change was made - what are we trying to do.

Content guidelines:
- Explain how the change affects the user - what is the new or modified behavior
- If the change affects performance, include measurements and description of how we measured
- If the change modifies the output, include example output with and without the change
- If the change introduces new logs, show example logs including the changed or new logs
- If several alternatives were considered, explain why we chose the particular solution
- Discuss the negative effects of the change if any
- If the change includes new APIs, describe the new APIs and how they are used
- Avoid describing details that are best seen in the diff

Footer:
- Include `Assisted-by: Cursor/{model-name}` footer
