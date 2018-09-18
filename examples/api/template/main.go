package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/circonus-labs/cosi-server/api"
	"github.com/pkg/errors"
)

func main() {
	osType := flag.String("type", "Linux", "OS type - e.g. Linux, Solaris, BSD, etc.")
	osDist := flag.String("dist", "CentOS", "OS distro - e.g. CentOS, Ubuntu, OmniOS, etc.")
	osVers := flag.String("vers", "7.4.1408", "OS version - e.g. 7.4.1408, 16.04, r151014, etc.")
	sysArch := flag.String("arch", "x86_64", "System architecture - e.g. x86_64, amd64, etc.")
	cosiURL := flag.String("url", "https://onestep.circonus.com", "COSI Server URL")
	id := flag.String("id", "check-system", "Template ID - e.g. graph-cpu, worksheet-system, graph-vm, etc.")
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

	if err := showTemplate(client, id); err != nil {
		log.Fatal(err)
	}
}

func showTemplate(c *api.Client, id *string) error {
	if *id == "" {
		return errors.New("invalid id (empty)")
	}
	parts := strings.Split(*id, "-")
	if len(parts) != 2 {
		return errors.Errorf("invalid id format (%s)", *id)
	}

	t, err := c.FetchTemplate(*id)
	if err != nil {
		return err
	}

	tcfg, err := json.MarshalIndent(t.Config, "", "    ")
	if err != nil {
		return errors.Wrap(err, "formatting template config")
	}

	fmt.Printf("COSI Template for id=%s\n", *id)

	fmt.Printf("Type       : %s\n", t.Type)
	fmt.Printf("Name       : %s\n", t.Name)
	fmt.Printf("Version    : %s\n", t.Version)
	fmt.Printf("Description: %s\n", t.Description)
	if len(t.Notes) > 0 {
		fmt.Println("Notes      :")
		for _, note := range t.Notes {
			fmt.Println("\t", note)
		}
	}

	if t.Type == "graph" {
		fmt.Printf("Variable   : %v\n", t.VariableMetrics)
		if t.VariableMetrics {
			fmt.Println("Filters    :")
			if len(t.Filters.Include) > 0 {
				fmt.Println("\tInclude:")
				for _, f := range t.Filters.Include {
					fmt.Println("\t\t", f)
				}
			}
			if len(t.Filters.Exclude) > 0 {
				fmt.Println("\tExclude:")
				for _, f := range t.Filters.Exclude {
					fmt.Println("\t\t", f)
				}
			}
		}
	}

	fmt.Println("Config(s)  :", string(tcfg))

	return nil
}
