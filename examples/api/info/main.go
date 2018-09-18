package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/circonus-labs/cosi-server/api"
)

func main() {
	// NOTE: these four parameters are not used for the info request, it would
	//       not make sense to require valid OS parameters in order to get a
	//       list of the supported operating systems...
	// osType := flag.String("type", "Linux", "OS type - e.g. Linux, Solaris, BSD, etc.")
	// osDist := flag.String("dist", "CentOS", "OS distro - e.g. CentOS, Ubuntu, OmniOS, etc.")
	// osVers := flag.String("vers", "7.4.1408", "OS version - e.g. 7.4.1408, 16.04, r151014, etc.")
	// sysArch := flag.String("arch", "x86_64", "System architecture - e.g. x86_64, amd64, etc.")
	cosiURL := flag.String("url", "https://onestep.circonus.com", "COSI Server URL")
	flag.Parse()

	cfg := api.Config{
		OSType:    "n/a", //*osType,
		OSDistro:  "n/a", // *osDist,
		OSVersion: "n/a", // *osVers,
		SysArch:   "n/a", // *sysArch,
		CosiURL:   *cosiURL,
	}

	client, err := api.New(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := showInfo(client); err != nil {
		log.Fatal(err)
	}
}

func showInfo(c *api.Client) error {
	info, err := c.FetchInfo()
	if err != nil {
		return err
	}

	fmt.Println("COSI Server Info")
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Printf("Version: %s\n", info.Version)

	fmt.Println("Supported operating systems:")
	for _, os := range info.Supported {
		fmt.Printf("\t%s\n", os)
	}

	return nil
}
