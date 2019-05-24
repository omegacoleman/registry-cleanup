package main

import (
	"os"
	"log"
	"fmt"
	"flag"
	"./config"
	"./filter"
	"./registry"
)

var (
	config_file_name string
	dry_run bool
	print_help bool
)

func init() {
	flag.BoolVar(&print_help, "h", false, "print help message")
	flag.StringVar(&config_file_name, "c", "./cleanup.yml", "config file name and path")
	flag.BoolVar(&dry_run, "n", false, "dry run, do not actually delete")
}

func main() {
	flag.Parse()

	if print_help {
		flag.PrintDefaults()
	}
	if env_config_file, ok := os.LookupEnv("REGISTRY_CLEANUP_CONFIG"); ok {
		config_file_name = env_config_file
	}

	data, err := config.LoadConfigFile(config_file_name)
	if err != nil {
		log.Fatalf("exited due to error while loading config file %v", err)
	}

	curr_config, err := config.Parse(data)
	if err != nil {
		log.Fatalf("exited due to error while parsing config file %v", err)
	}

	log.Printf("Here we go : %v\n", curr_config)

	curr_registry := registry.New(curr_config.RegistryURL)

	if curr_config.Auth != nil {
		curr_registry.SetBasicAuth(curr_config.Auth.Username, curr_config.Auth.Password)
	}

	err = curr_registry.CheckAPIVersionV2()
	if err != nil {
		log.Fatalf("error checking api version : %v", err)
	}
	log.Println("Checked /v2/, server is alive")

	skipped := []string{}
	for _, curr_repo := range curr_config.RepositoryArray {
		log.Println("Working on repo " + curr_repo + " ...")

		curr_tags, err := curr_registry.ListImageTags(curr_repo)
		if err != nil {
			log.Fatalf("error listing tags : %v", err)
		}

		m, err := filter.MultiFilter(curr_config.MatchRules, curr_config.ExceptRules, curr_tags)
		if err != nil {
			log.Fatalf("exited due to error while filtering tags %v", err)
		}

		del_tags := make([]string, 0, len(*m))

		log.Println("Tags marked for deletion : ")
		for ele, is_set := range *m {
			if is_set {
				del_tags = append(del_tags, ele)
				fmt.Println(" * " + ele)
			}
		}

		log.Println("Deleting tags above...")
		for _, curr_tag := range del_tags {
			curr_digest, err := curr_registry.CheckImageManifest(curr_repo, curr_tag)
			if err != nil {
				log.Printf("error getting manifest digest : %v, skipping ..", err)
				skipped = append(skipped, curr_repo + "/" + curr_tag)
				continue
			}
			log.Println("Got digest for " + curr_tag + ":" + curr_digest)

			log.Println("Deleting ...")
			if ! dry_run {
				err = curr_registry.DeleteImage(curr_repo, curr_digest)
				if err != nil {
					log.Printf("error deleting image: %v, skipping ..", err)
					skipped = append(skipped, curr_repo + "/" + curr_tag)
					continue
				}
			}
		}
	}

	log.Println("Done!")

	if len(skipped) > 0 {
		log.Println("These tags encountered problems while deleting : ")
		for _, e := range skipped {
			log.Println(" * " + e)
		}
		log.Println("If there are still problems after rerun, you may have to check them manually.")
	}
}

