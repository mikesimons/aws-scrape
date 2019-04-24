package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ryanuber/go-glob"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/sync/semaphore"
)

type recordFn func(*session.Session, string, string) []Record

type Record struct {
	File  string      `json:"file"`
	Attrs interface{} `json:"attrs,omitempty"`
}

type AccountConfig struct {
	Role    string `yaml:"role"`
	Profile string `yaml:"profile"`
}

type Scrape struct {
	AccountIDs []string `yaml:"account_ids"`
	Role       string   `yaml:"role"`
	Regions    []string `yaml:"regions"`
	Resources  []string `yaml:"resources"`
}

var versionString = "dev"
var resourceMap map[string]recordFn

func addResource(name string, fn recordFn) {
	if resourceMap == nil {
		resourceMap = make(map[string]recordFn)
	}
	resourceMap[name] = fn
}

func main() {
	var maxConcurrent int64
	var accountConfigString string
	var scrapeConfigStrings []string

	var accountConfigs map[string]AccountConfig
	var scrapeConfigs []Scrape

	cmd := &cobra.Command{
		Use:     "aws-scrape",
		Version: versionString,
		Run: func(cmd *cobra.Command, args []string) {
			out := make(chan Record)
			go func() {
				var allRegions []string
				s := awsSession("", "", "")
				regionRecords := regions(s, "", "")
				for _, r := range regionRecords {
					allRegions = append(allRegions, r.Attrs.(map[string]interface{})["name"].(string))
				}

				accountConfigString = fmt.Sprintf("{%s}", accountConfigString)
				err := yaml.Unmarshal([]byte(accountConfigString), &accountConfigs)
				if err != nil {
					panic(err)
				}

				for _, str := range scrapeConfigStrings {
					var tmp Scrape
					str := fmt.Sprintf("{%s}", str)
					err := yaml.Unmarshal([]byte(str), &tmp)
					if err != nil {
						panic(err)
					}
					scrapeConfigs = append(scrapeConfigs, tmp)
				}

				ctx := context.TODO()
				sem := semaphore.NewWeighted(maxConcurrent)
				for _, config := range scrapeConfigs {
					for _, accountID := range config.AccountIDs {
						regions := filterRegions(allRegions, config.Regions)
						for _, region := range regions {
							session := awsSession(region, accountConfigs[accountID].Profile, fmt.Sprintf(accountConfigs[accountID].Role, accountID))
							for _, resource := range config.Resources {
								if err := sem.Acquire(ctx, 1); err != nil {
									log.Printf("Failed to acquire semaphore: %v", err)
									break
								}

								go func(resource string) {
									defer sem.Release(1)
									fn, ok := resourceMap[resource]
									if !ok {
										log.Fatalf("Invalid resource: %s\n", resource)
									}
									for _, r := range fn(session, region, accountID) {
										out <- r
									}
								}(resource)
							}
						}
					}
				}

				if err := sem.Acquire(ctx, maxConcurrent); err != nil {
					log.Printf("Failed to acquire semaphore: %v", err)
				}
				close(out)
			}()

			for v := range out {
				out, _ := json.Marshal(v)
				fmt.Printf("%s\n", out)
			}
		},
	}

	cmd.Flags().StringVarP(&accountConfigString, "accounts", "a", accountConfigString, "")
	cmd.Flags().StringArrayVarP(&scrapeConfigStrings, "scrape", "s", scrapeConfigStrings, "")
	cmd.Flags().Int64VarP(&maxConcurrent, "max-concurrent", "m", 10, "Maximum concurrent resource goroutines")

	err := cmd.Execute()
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}

func filterRegions(all []string, requested []string) []string {
	var output []string
	for _, region := range all {
		for _, input := range requested {
			if glob.Glob(input, region) {
				output = append(output, region)
			}
		}
	}
	return output
}

func getTagOrDefault(tags []*ec2.Tag, name string, def string) string {
	for _, t := range tags {
		if *t.Key == name {
			return *t.Value
		}
	}
	return def
}

func awsSession(region string, profile string, role string) *session.Session {
	options := session.Options{
		Config:            aws.Config{},
		SharedConfigState: session.SharedConfigEnable,
	}

	if region != "" {
		options.Config.Region = aws.String(region)
	}

	if profile != "" {
		options.Profile = profile
	}

	s := session.Must(session.NewSessionWithOptions(options))

	if role != "" {
		options.Config.Credentials = stscreds.NewCredentials(s, role, func(p *stscreds.AssumeRoleProvider) {})
		s = session.Must(session.NewSession(&options.Config))
	}

	return s
}
