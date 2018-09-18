package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/circonus-labs/cosi-server/api"
)

func main() {
	osType := flag.String("type", "Linux", "OS type - e.g. Linux, Solaris, BSD, etc.")
	osDist := flag.String("dist", "CentOS", "OS distro - e.g. CentOS, Ubuntu, OmniOS, etc.")
	osVers := flag.String("vers", "7.4.1408", "OS version - e.g. 7.4.1408, 16.04, r151014, etc.")
	sysArch := flag.String("arch", "x86_64", "System architecture - e.g. x86_64, amd64, etc.")
	cosiURL := flag.String("url", "https://onestep.circonus.com", "COSI Server URL")
	flag.Parse()

	cfg := api.Config{
		OSType:    *osType,
		OSDistro:  *osDist,
		OSVersion: *osVers,
		SysArch:   *sysArch,
		CosiURL:   *cosiURL,
	}

	client, err := api.New(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := showPackage(client); err != nil {
		log.Fatal(err)
	}
}

func showPackage(c *api.Client) error {
	pkg, err := c.FetchPackage("json")
	if err != nil {
		return err
	}

	fmt.Println("COSI Agent Package")
	if pkg.File != "" {
		fmt.Printf("URL : %s\n", pkg.URL)
		fmt.Printf("File: %s\n", pkg.File)
	} else {
		fmt.Printf("Publisher: %s\n", pkg.PublisherName)
		fmt.Printf("Pub URL  : %s\n", pkg.PublisherURL)
		fmt.Printf("Package  : %s\n", pkg.Name)
	}

	return nil
}
