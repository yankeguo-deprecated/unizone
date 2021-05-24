package main

import (
	"context"
	"flag"
	"go.guoyk.net/unizone/pkg/providers"
	"log"
	"os"
)

func main() {
	var (
		optVerbose bool
		optConf    string
		optOutput  string
	)

	flag.BoolVar(&optVerbose, "v", false, "enable verbose logging")
	flag.StringVar(&optConf, "c", "unizone.yml", "configuration file")
	flag.StringVar(&optOutput, "o", "unizone.zone", "output dns zone file")
	flag.Parse()

	var err error
	defer func(err *error) {
		if *err != nil {
			log.Println("exited with error:", (*err).Error())
			os.Exit(1)
		} else {
			log.Println("exited")
		}
	}(&err)

	var cfg Config
	if err = LoadConfigFile(optConf, &cfg); err != nil {
		return
	}

	log.Println("configuration loaded")

	var records []providers.Record

	for _, cloud := range cfg.Providers {
		log.Println("inspecting provider:", cloud.ID)
		for _, network := range cloud.Networks {
			log.Println("inspecting network:", cloud.ID, network.Region, network.ID)
			var provider providers.Provider
			if provider, err = providers.Create(cloud.Provider, providers.Options{
				ID:          cloud.ID,
				TokenID:     cloud.TokenID,
				TokenSecret: cloud.TokenSecret,
				Region:      network.Region,
			}); err != nil {
				return
			}
			for _, service := range cloud.Services {
				log.Println("inspecting service:", cloud.ID, network.Region, network.ID, service)
				var cloudRecords []providers.Record
				if cloudRecords, err = provider.ListRecords(context.Background(), network.ID, service); err != nil {
					return
				}
			outerLoop1:
				for _, cloudRecord := range cloudRecords {
					for _, knownRecord := range records {
						if knownRecord.Name == cloudRecord.Name {
							log.Println(
								"found duplicated record:",
								cloud.ID,
								network.Region,
								network.ID,
								service,
								cloudRecord.Name,
							)
							continue outerLoop1
						}
					}
					if optVerbose {
						log.Println("found record:", cloud.ID, network.Region, network.ID, service, cloudRecord.Name)
					}
					records = append(records, cloudRecord)
				}
			}
		}
	}
}