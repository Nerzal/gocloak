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

// Version is a structure to represent a version
type Version struct {
	Major int
	Minor int
	Patch int
}

// String returns a string representation of version
func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Parse converts a string representation of version
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

// Less compares the v with another
func (v *Version) Less(another *Version) bool {
	return v.Major < another.Major || v.Minor < another.Minor || v.Patch < another.Patch
}

// ByVersion is an interface to sort versions
type ByVersion []Version

// Len returns a length of array
func (a ByVersion) Len() int { return len(a) }

// Swap swaps i and j items
func (a ByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less checks that an i element less than a j element
func (a ByVersion) Less(i, j int) bool {
	return a[i].Major < a[j].Major && a[i].Minor < a[j].Minor && a[i].Patch < a[j].Patch
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
	return strings.Trim(stdout.String(), "\n")
}

func getLastTags() Version {
	var tags []Version
	for _, tag := range strings.Split(git(nil, "tag", "--list"), "\n") {
		tag = strings.TrimSpace(tag)
		if len(tag) > 0 {
			var ver Version
			if ver.Parse(tag) {
				tags = append(tags, ver)
			}
		}
	}
	if len(tags) > 0 {
		sort.Sort(ByVersion(tags))
		return tags[len(tags)-1]
	}
	return Version{}
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

func bumpTag(tag Version, annotation string, sign bool, ref string) {
	input := strings.NewReader(annotation)
	args := []string{
		"tag",
	}
	if sign {
		args = append(args, "--sign")
	}
	args = append(args, "-F-", tag.String(), ref)
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
	printFlag("commit", "c")
	printFlag("last-tag", "t")
}

func createBoolFlag(name, short string, usage string) *bool {
	p := flag.Bool(name, false, usage)
	flag.BoolVar(p, short, false, "")
	return p
}

func createStrFlag(name, short string, value string, usage string) *string {
	p := flag.String(name, value, usage)
	flag.StringVar(p, short, value, "")
	return p
}

func main() {
	major := createBoolFlag("major", "m", "Increase major version")
	minor := createBoolFlag("minor", "n", "Increase minor version")
	patch := createBoolFlag("patch", "p", "Increase patch version")
	sign := createBoolFlag("sign", "s", "Make a GPG-signed tag, using the default e-mail address's key")
	dryRun := createBoolFlag("dry-run", "r", "Test run, prints an annotation of the tag")
	ref := createStrFlag("commit", "c", "HEAD", "The commit that the new tag will refer to")
	tag := createStrFlag("last-tag", "t", "", "The last tag that the new tag will compare to")

	flag.Usage = printUsage
	flag.Parse()
	args := flag.Args()
	var lastTag Version
	if len(*tag) > 0 {
		if !lastTag.Parse(*tag) {
			fmt.Printf("incorrect tag: %s\n", *tag)
			os.Exit(1)
		}
	} else {
		lastTag = getLastTags()
	}
	var newTag Version
	if len(args) > 0 {
		if !newTag.Parse(args[0]) {
			fmt.Printf("incorrect version: %s\n", args[0])
			os.Exit(1)
		}
	} else {
		newTag = lastTag

		if *major {
			newTag.Major++
			newTag.Minor = 0
			newTag.Patch = 0
		}
		if *minor || (!*major && !*patch) {
			newTag.Minor++
			newTag.Patch = 0
		}
		if *patch {
			newTag.Patch++
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
			*ref,
		)
		showNewTag(newTag)
	}
}
