# Internal application code

This folder is reserved for Brocode's implementation as we build it.
Go prevents projects outside Brocode from importing packages under `internal/`.

Start with code in `main.go`. When a responsibility grows enough to deserve its
own package, move it here—for example, `internal/harness` for the loop that
coordinates model requests and software engineering tools.

We will add packages when they are useful, rather than defining them all upfront.
