package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "usage: invoice-service NUMBER CUSTOMER EMAIL AMOUNT")
		os.Exit(2)
	}
	c, err := NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	id, err := IssueInvoice(context.Background(), c, Invoice{Number: os.Args[1], Customer: os.Args[2], Email: os.Args[3], Amount: os.Args[4]})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(map[string]string{"message_id": id})
}
