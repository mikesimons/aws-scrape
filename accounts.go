package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/organizations"

	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("accounts", accounts)
}

func accounts(s *session.Session, _ string, _ string) []Record {
	fmt.Fprintf(os.Stderr, "Loading account list\n")
	svc := organizations.New(s)
	input := &organizations.ListAccountsInput{}
	result, err := svc.ListAccounts(input)
	if err != nil {
		log.Fatalf("ListAccount error: %s", err)
	}

	var output []Record
	for _, a := range result.Accounts {
		output = append(output, Record{
			File: "aws-accounts",
			Attrs: map[string]interface{}{
				"name":       *a.Name,
				"email":      *a.Email,
				"account_id": *a.Id,
				"arn":        *a.Arn,
			},
		})
	}
	return output
}
