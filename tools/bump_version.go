package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) Parse(raw string) bool {
	v.Major = 0
	v.Minor = 0
	v.Patch = 0
	if raw == "" {
		return true
	}
	if raw[0] != 'v' {
		return false
	}
	parts := strings.Split(raw[1:], ".")
	p, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	v.Major = p
	if len(parts) > 1 {
		p, err := strconv.Atoi(parts[1])
		if err != nil {
			return false
		}
		v.Minor = p
	}
	if len(parts) > 2 {
		p, err := strconv.Atoi(parts[2])
		if err != nil {
			return false
		}
		v.Patch = p
	}

	return true
}

func (v *Version) Less(another *Version) bool {
	return v.Major < another.Major || v.Minor < another.Minor || v.Patch < another.Patch
}

func git(input *strings.Reader, arg ...string) string {
	cmd := exec.Command("git", arg...)
	if input != nil {
		cmd.Stdin = input
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Print(cmd.Args)
		log.Fatal(err)
	}
	if stderr.Len() > 0 {
		log.Fatal(stderr.String())
	}
	return stdout.String()
}

func getLastTags() Version {
	var tags []*Version
	for _, tag := range strings.Split(git(nil, "tag", "--list"), "\n") {
		var ver Version
		if ver.Parse(tag) {
			tags = append(tags, &ver)
		}
	}
	sort.Slice(tags, func(i, j int) bool {
		return !tags[i].Less(tags[j])
	})
	return *tags[0]
}

func getChangeLog(version Version) []string {
	var res []string
	for _, line := range strings.Split(git(
		nil,
		"log",
		"--pretty=* %h [%s](http://github.com/Nerzal/gocloak/commit/%H)",
		"--no-merges",
		"--reverse",
		fmt.Sprintf("%s..HEAD", version.String()),
	), "\n") {
		line = strings.Trim(line, "\n ")
		//if len(line) > 0 && !strings.HasPrefix(line, "gpg:") && !strings.HasPrefix(line, "Primary key fingerprint:") {
		if len(line) > 0 && strings.HasPrefix(line, "*") {
			res = append(res, line)
		}
	}
	return res
}

func makeAnnotation(tag Version, changeLog []string) string {
	subject := fmt.Sprintf("Bump version %s", tag.String())
	sep := strings.Repeat("-", len(subject))
	annotation := []string{
		subject,
		sep,
		"",
	}
	annotation = append(annotation, changeLog...)
	return strings.Join(annotation, "\n")
}

func bumpTag(tag Version, annotation string, sign bool) {
	input := strings.NewReader(annotation)
	args := []string{
		"tag",
	}
	if sign {
		args = append(args, "--sign")
	}
	args = append(args, "-F-", tag.String())
	_ = git(
		input,
		args...,
	)
}

func showNewTag(tag Version) {
	out := git(
		nil,
		"show",
		tag.String(),
	)
	println(out)
}

func printFlag(name, short string) {
	f := flag.Lookup(name)
	fmt.Printf("    -%s, --%s\n", short, name)
	println("       ", f.Usage)
}

func printUsage() {
	println("bump_version [options] [version]")
	println("Options:")

	printFlag("major", "m")
	printFlag("minor", "n")
	printFlag("patch", "p")
	printFlag("sign", "s")
	printFlag("dry-run", "r")
}

func createFlag(name, short string, usage string) *bool {
	p := flag.Bool(name, false, usage)
	flag.BoolVar(p, short, false, "")
	return p
}

func main() {
	major := createFlag("major", "m", "Increase major version")
	minor := createFlag("minor", "n", "Increase minor version")
	patch := createFlag("patch", "p", "Increase patch version")
	sign := createFlag("sign", "s", "Make a GPG-signed tag, using the default e-mail address's key")
	dryRun := createFlag("dry-run", "r", "Test run, prints an annotation of the tag")

	flag.Usage = printUsage
	flag.Parse()
	args := flag.Args()
	lastTag := getLastTags()
	var newTag Version
	if len(args) > 0 {
		if !newTag.Parse(args[0]) {
			fmt.Printf("incorrect version: %s\n", args[0])
			os.Exit(1)
		}
	} else {
		newTag = lastTag

		if *major {
			newTag.Major += 1
		}
		if *minor || (!*major && !*patch) {
			newTag.Minor += 1
		}
		if *patch {
			newTag.Patch += 1
		}
	}
	changeLog := getChangeLog(lastTag)
	annotation := makeAnnotation(newTag, changeLog)
	if *dryRun {
		println(annotation)
	} else {
		bumpTag(
			newTag,
			annotation,
			*sign,
		)
		showNewTag(newTag)
	}
}
