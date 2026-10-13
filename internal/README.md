# Internal application code

This folder contains Brocode's implementation as we build it.
Go prevents projects outside Brocode from importing packages under `internal/`.

`openrouter/` contains the text chat client. `sandbox/` is a placeholder for
command execution. `main.go` connects application packages to the command line.

When another responsibility grows enough to deserve its own package, move it
here—for example, `internal/harness` for the loop that coordinates model
requests and software engineering tools.

We will add packages when they are useful, rather than defining them all upfront.
