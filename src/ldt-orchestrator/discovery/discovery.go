package discovery

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"longevity/src/ldt-orchestrator/github"
	. "longevity/src/types"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strings"
)

type DiscoveryConfig struct {
	repository_source string
	SupportedLDTs     *LDTList
	otherLDTs         *LDTList
	repositories      []string
	os                string
	arch              string
}

func NewConfig(path string) *DiscoveryConfig {
	os, arch := getRuntimeInformation()
	return &DiscoveryConfig{
		repository_source: path,
		SupportedLDTs:     NewLDTList(),
		otherLDTs:         NewLDTList(),
		repositories:      make([]string, 0),
		os:                os,
		arch:              arch,
	}
}

func getRuntimeInformation() (string, string) {
	os := runtime.GOOS
	arch := runtime.GOARCH
	return os, arch
}

func (c *DiscoveryConfig) DiscoverLDTs() {
	c.updateRepositories()
	newLDTs := github.FetchGithubReleases(c.repositories)
	for _, ldt := range newLDTs.LDTs {
		if ldt.Os == c.os && ldt.Arch == c.arch {
			if !ldtAlreadyExists(&ldt, c.SupportedLDTs) {
				c.SupportedLDTs.LDTs = append(c.SupportedLDTs.LDTs, ldt)
			}
		} else {
			if !ldtAlreadyExists(&ldt, c.SupportedLDTs) {
				c.otherLDTs.LDTs = append(c.otherLDTs.LDTs, ldt)
			}
		}
	}
	c.addLatestTag()
}

func (c *DiscoveryConfig) addLatestTag() {
	sort.Slice(c.SupportedLDTs.LDTs, func(i, j int) bool {
		return c.SupportedLDTs.LDTs[i].Version > c.SupportedLDTs.LDTs[j].Version
	})
	fmt.Println(c.SupportedLDTs)
}

func (c *DiscoveryConfig) GetUrlFromLDT(id int) (string, error) {
	if id >= len(c.SupportedLDTs.LDTs) {
		return "", errors.New("Failed to map ID to LDT")
	}
	return c.SupportedLDTs.LDTs[id].Url, nil
}

func ldtAlreadyExists(ldt *LDT, ldt_list *LDTList) bool {
	for _, existingLDT := range ldt_list.LDTs {
		if string(ldt.Hash) == string(existingLDT.Hash) {
			return true
		}
	}
	return false
}

func isGithubRepository(repo string) bool {
	stuff, _ := url.Parse(repo)
	if stuff.Host != "" {
		return strings.Contains(stuff.Host, "github.com")
	} else if strings.HasPrefix(stuff.Path, "www.github.com") {
		return true
	} else if strings.HasPrefix(stuff.Path, "github.com") {
		return true
	} else {
		return false
	}
}

func (c *DiscoveryConfig) updateRepositories() {
	var content io.Reader
	if isURL(c.repository_source) {
		resp, _ := http.Get(c.repository_source)
		defer resp.Body.Close()
		content = resp.Body
	} else {
		file, err := os.Open(c.repository_source)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		content = file
	}
	scanner := bufio.NewScanner(content)
	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "#") {
			c.repositories = append(c.repositories, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func isURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}
