// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// babble tangles lit books without Emacs. This file is tangled from
// BOOK.org.
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bilus/babble/internal/book"
	"github.com/bilus/babble/internal/org"
	"github.com/bilus/babble/internal/tangle"
)

const usage = `usage: babble <command> [flags] [BOOK.org]

commands:
  tangle  write every eligible code block to its target file
  parse   read the book and report whether it parses (--dump for JSON)
  init    start a lit book in an empty directory (--module, --dir)
  vendor  write the embedded lit/ into this directory (--dir)
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches the subcommands and owns the exit codes.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "parse":
		fs := flag.NewFlagSet("parse", flag.ContinueOnError)
		fs.SetOutput(stderr)
		dump := fs.Bool("dump", false, "print the tree as JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() != 1 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		d, err := org.Parse(fs.Arg(0))
		if err == nil {
			err = org.ResolveAll(d)
		}
		if err != nil {
			fmt.Fprintln(stderr, "babble:", err)
			return 1
		}
		if *dump {
			out, err := book.Dump(d)
			if err != nil {
				fmt.Fprintln(stderr, "babble:", err)
				return 1
			}
			if _, err := stdout.Write(out); err != nil {
				return 1
			}
		}
		return 0
	case "tangle":
		if len(args) != 2 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		return tangleCmd(args[1], stderr)
	case "init":
		fs := flag.NewFlagSet("init", flag.ContinueOnError)
		fs.SetOutput(stderr)
		module := fs.String("module", "", "go module path (default: the directory name)")
		dir := fs.String("dir", ".", "directory to set up")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() != 0 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		return initCmd(*dir, *module, stdout, stderr)
	case "vendor":
		fs := flag.NewFlagSet("vendor", flag.ContinueOnError)
		fs.SetOutput(stderr)
		dir := fs.String("dir", ".", "directory to write lit/ into")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() != 0 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		return vendorCmd(*dir, stdout, stderr)
	}
	fmt.Fprint(stderr, usage)
	return 2
}

// The walk writes only what differs, so the report names the files that
// were behind and says nothing when the copy is already current. A
// directory that is not there is refused rather than created: init
// makes a project and this refreshes one, so a mistyped path should
// say so instead of leaving a directory holding nothing but a
// toolchain. The toolchain's own checkout is refused too, since there
// the embedded copy is whatever was compiled in and the source beside
// it may be newer.
func vendorLit(dir string, out io.Writer) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, "lit", "BOOK.org")); err == nil {
		return fmt.Errorf("%s holds lit/BOOK.org and is the toolchain's own checkout; run make tangle in lit/ instead", dir)
	}
	wrote := 0
	err = fs.WalkDir(litFS, "lit", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		want, err := litFS.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(dir, filepath.FromSlash(path))
		if old, err := os.ReadFile(target); err == nil && bytes.Equal(old, want) {
			return nil
		}
		if err := replaceFile(target, want); err != nil {
			return err
		}
		wrote++
		fmt.Fprintln(out, path)
		return nil
	})
	if err == nil && wrote == 0 {
		fmt.Fprintln(out, "lit/ is already current")
	}
	return err
}

// A file is written beside its target and renamed over it. The rename
// is what makes the file whole-old or whole-new, so a run that fails
// partway leaves no mixture of two toolchains, and it is also what
// replaces a symbolic link instead of writing through it to whatever
// it points at.
func replaceFile(target string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".babble-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
}

// Two more steps belong here and are not here yet. [[#stage-9][Stage 9]] adds a
// refresh of the book's dynamic blocks before the tangle, and a
// reparse after it, because refreshing rewrites the book and moves
// every line number the tangle depends on. [[#stage-10][Stage 10]] adds the lint
// pass, which is the half of the fence that needs the whole document.
// Until both land this function is three calls, and the code below is
// what runs.
func tangleCmd(bookPath string, stderr io.Writer) int {
	d, err := org.Parse(bookPath)
	if err == nil {
		err = org.ResolveAll(d)
	}
	if err == nil {
		err = org.LintQuotedDelimiters(d)
	}
	if err == nil {
		err = tangle.Run(d)
	}
	if err != nil {
		fmt.Fprintln(stderr, "babble:", err)
		return 1
	}
	return 0
}

// Each file is replaced whole or not at all, written beside its target
// and renamed over it, so a run that fails partway leaves no mixture of
// two toolchains and a symlink under lit/ is replaced rather than
// written through. Nothing is removed, because several paths sit
// outside the embedded set on purpose and the repository holding them
// is the toolchain's own; sweeping lit/ clean would delete the source
// the rest is tangled from. Nothing outside lit/ is touched at all,
// which is what lets a build call it without knowing what is already
// there.
func vendorCmd(dir string, stdout, stderr io.Writer) int {
	if err := vendorLit(dir, stdout); err != nil {
		fmt.Fprintln(stderr, "babble:", err)
		return 1
	}
	return 0
}

// Two files are left out on purpose. lit/BOOK.org is lit's own
// literate source and a consumer never tangles it, and lit/example
// holds a symbolic link, which embedding cannot carry.
//go:embed lit/lit.mk lit/tangle.el lit/preprocess.py lit/texlinks.py
//go:embed lit/mermaid.lua lit/notes.lua lit/plan.lua lit/svg.lua
//go:embed lit/README.md lit/AGENTS.md lit/templates
var litFS embed.FS

// The tool stack is pinned in devbox.json rather than assumed. Every
// one of these is reached by lit.mk or by a filter it runs: emacs
// tangles, python3 runs the preprocessor and the link fixer, pandoc
// and tectonic make the PDF, mmdc draws the mermaid diagrams, and
// rsvg-convert turns a hand-drawn SVG figure into a PDF the LaTeX can
// include.
const devboxJSON = `{
  "packages": [
    "go@latest",
    "gnumake@latest",
    "git@latest",
    "emacs-nox@latest",
    "python@latest",
    "pandoc@latest",
    "tectonic@latest",
    "mermaid-cli@latest",
    "librsvg@latest"
  ]
}
`

const gitignore = `book-weave.*
.devbox/
.venv/
`

// initCmd is the whole setup in the order the steps depend on each
// other. The environment comes first, because make setup runs emacs
// and emacs is one of the packages. Then the toolchain, the Makefile
// and go.mod, because make setup reads all three: it picks its
// template from go.mod, and it refuses to run at all if a BOOK.org is
// already there. Then the book is written and tangled, and last it is
// woven, which is the step that proves the PDF path works on this
// machine.
func initCmd(dir, module string, stdout, stderr io.Writer) int {
	steps := []struct {
		what string
		run  func(*project) error
	}{
		{"writing devbox.json", (*project).writeDevbox},
		{"generating .envrc", (*project).writeEnvrc},
		{"allowing direnv", (*project).allowDirenv},
		{"copying lit", (*project).copyLit},
		{"writing Makefile and .gitignore", (*project).writeMakefile},
		{"running go mod init", (*project).goModInit},
		{"running make setup", (*project).makeSetup},
		{"running make weave", (*project).makeWeave},
	}
	p, err := newProject(dir, module, stdout)
	if err != nil {
		fmt.Fprintln(stderr, "babble:", err)
		return 1
	}
	for i, step := range steps {
		fmt.Fprintf(stdout, "[%d/%d] %s\n", i+1, len(steps), step.what)
		if err := step.run(p); err != nil {
			fmt.Fprintln(stderr, "babble:", err)
			return 1
		}
	}
	fmt.Fprintf(stdout, "\ndone. %s holds a book, its sources, and book.pdf.\n", p.dir)
	fmt.Fprintln(stdout, "edit BOOK.org, then run make tangle.")
	return 0
}

// A project is the directory being set up plus the two facts every
// step needs: what to call the Go module, and where progress goes.
// newProject is where the refusals live, so no step has to check
// whether it is safe to write.
type project struct {
	dir    string
	module string
	out    io.Writer
	devbox bool // devbox is on PATH, so commands run inside it
}

func newProject(dir, module string, out io.Writer) (*project, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	for _, taken := range []string{"BOOK.org", "lit", "go.mod"} {
		if _, err := os.Stat(filepath.Join(dir, taken)); err == nil {
			return nil, fmt.Errorf("%s already has %s; init needs an empty directory", dir, taken)
		}
	}
	if module == "" {
		module = filepath.Base(abs)
	}
	_, hasDevbox := exec.LookPath("devbox")
	return &project{dir: dir, module: module, out: out, devbox: hasDevbox == nil}, nil
}

// The two writers are plain, and both refuse to overwrite. Nothing
// here should ever clobber a file a person wrote, and newProject has
// already checked the three that matter, so a collision this deep is
// a bug rather than a user error.
func (p *project) write(name, content string) error {
	path := filepath.Join(p.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, content)
	return err
}

func (p *project) writeDevbox() error { return p.write("devbox.json", devboxJSON) }

func (p *project) writeMakefile() error {
	if err := p.write("Makefile", "include lit/lit.mk\n"); err != nil {
		return err
	}
	return p.write(".gitignore", gitignore)
}

// copyLit walks the embedded files and writes them out under lit/,
// stripping nothing: the paths inside the archive already start with
// lit/, because that is where they sit in this repository.
func (p *project) copyLit() error {
	return fs.WalkDir(litFS, "lit", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, err := litFS.ReadFile(path)
		if err != nil {
			return err
		}
		return p.write(path, string(body))
	})
}

// Without devbox there is no .envrc and nothing fetches the tools, so
// the least useful thing to print is that a step was skipped. What a
// reader needs is the list, and devbox.json already is the list.
// Parsing it back out is what keeps the message from drifting away
// from the manifest: adding a package to the stack updates both.
func (p *project) reportStack() error {
	var cfg struct{ Packages []string }
	if err := json.Unmarshal([]byte(devboxJSON), &cfg); err != nil {
		return err
	}
	fmt.Fprintln(p.out, "  devbox is not installed, so no .envrc was written and")
	fmt.Fprintln(p.out, "  nothing will fetch the tool stack. devbox.json lists it:")
	for _, pkg := range cfg.Packages {
		name, _, _ := strings.Cut(pkg, "@")
		fmt.Fprintln(p.out, "    -", name)
	}
	fmt.Fprintln(p.out, "  install those, or install devbox and run:")
	fmt.Fprintln(p.out, "    devbox generate direnv && direnv allow")
	return nil
}

// Two of the steps can find nothing to do. Allowing direnv needs an
// .envrc, which only exists if devbox wrote one, and it needs direnv
// itself; either way the step says what is missing and the setup goes
// on. A step that skips is not a step that failed.
func (p *project) run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = p.dir
	cmd.Stdout, cmd.Stderr = p.out, p.out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func (p *project) runInStack(name string, args ...string) error {
	if p.devbox {
		return p.run("devbox", append([]string{"run", "--", name}, args...)...)
	}
	return p.run(name, args...)
}

func (p *project) writeEnvrc() error {
	if !p.devbox {
		p.reportStack()
		return nil
	}
	return p.run("devbox", "generate", "direnv", "--force")
}

func (p *project) allowDirenv() error {
	if _, err := os.Stat(filepath.Join(p.dir, ".envrc")); err != nil {
		fmt.Fprintln(p.out, "  no .envrc to allow")
		return nil
	}
	if _, err := exec.LookPath("direnv"); err != nil {
		fmt.Fprintln(p.out, "  direnv is not installed; the .envrc will do nothing until it is")
		return nil
	}
	return p.run("direnv", "allow")
}

func (p *project) goModInit() error { return p.run("go", "mod", "init", p.module) }
func (p *project) makeSetup() error { return p.runInStack("make", "setup") }
func (p *project) makeWeave() error { return p.runInStack("make", "weave") }
