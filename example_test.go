package awsysco_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

// Example functions in this file demonstrate core SDK usage. None of them
// carry an "Output:" comment, so `go test` compiles and type-checks them
// (proving the calls match the real, current method signatures) but does
// not execute them — running them for real would require a live AWSYS.CO
// server and a valid API key.

// Example demonstrates creating a client and shortening a URL.
func Example() {
	client := awsysco.NewClient("awsys_your_api_key_here")

	link, err := client.Links.Create(context.Background(), awsysco.CreateLinkInput{
		URL:        "https://example.com/very/long/url",
		CustomSlug: "my-slug",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Short URL:", link.ShortURL)
}

// Example_configuration shows the functional options available when
// constructing a client, and the AWSYS_API_KEY / AWSYS_BASE_URL environment
// variable fallbacks.
func Example_configuration() {
	// Explicit configuration.
	client := awsysco.NewClient("awsys_your_api_key_here",
		awsysco.WithBaseURL("https://staging.awsys.co"),
		awsysco.WithTimeout(15*time.Second),
		awsysco.WithMaxRetries(5),
	)
	_ = client

	// Or fall back to AWSYS_API_KEY / AWSYS_BASE_URL from the environment.
	envClient := awsysco.NewClient("")
	_ = envClient
}

// ExampleLinksResource_Create shows creating a link with optional fields set.
func ExampleLinksResource_Create() {
	client := awsysco.NewClient("awsys_your_api_key_here")

	maxClicks := 1000
	link, err := client.Links.Create(context.Background(), awsysco.CreateLinkInput{
		URL:       "https://example.com/product/123",
		MaxClicks: &maxClicks,
		Tags:      []string{"campaign", "q3"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(link.ShortURL)
}

// ExampleLinksResource_List shows retrieving a single page of links.
func ExampleLinksResource_List() {
	client := awsysco.NewClient("awsys_your_api_key_here")

	resp, err := client.Links.List(context.Background(), awsysco.ListLinksInput{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range resp.Links {
		fmt.Println(l.ShortURL, "->", l.Long)
	}
}

// ExampleLinksResource_Iter shows using the pagination iterator to walk every
// link matching a query without manually tracking offsets.
func ExampleLinksResource_Iter() {
	client := awsysco.NewClient("awsys_your_api_key_here")

	it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{Limit: 50})
	for it.Next() {
		link := it.Link()
		fmt.Println(link.ShortURL)
	}
	if err := it.Err(); err != nil {
		log.Fatal(err)
	}
}

// Example_errorHandling shows inspecting an SDK error with the Is*
// predicates and extracting the underlying *awsysco.AwsysError.
func Example_errorHandling() {
	client := awsysco.NewClient("awsys_your_api_key_here")

	_, err := client.Links.Get(context.Background(), "nonexistent_id")
	if err == nil {
		return
	}

	switch {
	case awsysco.IsNotFound(err):
		fmt.Println("link not found")
	case awsysco.IsAuthError(err):
		fmt.Println("invalid or expired API key")
	case awsysco.IsRateLimitError(err):
		var rlErr *awsysco.RateLimitError
		if errors.As(err, &rlErr) {
			fmt.Println("rate limited, retry after:", rlErr.RetryAfter)
		}
	case awsysco.IsServerError(err):
		fmt.Println("platform 5xx error")
	default:
		var awsysErr *awsysco.AwsysError
		if errors.As(err, &awsysErr) {
			fmt.Println("unexpected error:", awsysErr.Status, awsysErr.Code)
		}
	}
}

// Example_configurationError shows that a misconfigured client (missing API
// key, invalid base URL) fails fast with a *awsysco.ConfigurationError,
// before any network call is attempted.
func Example_configurationError() {
	client := awsysco.NewClient("", awsysco.WithBaseURL("ftp://example.com"))

	_, err := client.Links.Get(context.Background(), "abc123")
	var cfgErr *awsysco.ConfigurationError
	if errors.As(err, &cfgErr) {
		fmt.Println("misconfigured client:", cfgErr)
	}
}

// ExampleImportsResource_WaitForCompletion shows starting a provider import
// and blocking until it reaches a terminal state.
func ExampleImportsResource_WaitForCompletion() {
	client := awsysco.NewClient("awsys_your_api_key_here")
	ctx := context.Background()

	job, err := client.Imports.Start(ctx, awsysco.ImportStartOptions{
		Provider:    "bitly",
		AccessToken: "bitly_access_token",
	})
	if err != nil {
		log.Fatal(err)
	}

	final, err := client.Imports.WaitForCompletion(ctx, job.ID, &awsysco.WaitOptions{
		PollInterval: 5 * time.Second,
		Timeout:      10 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("import finished with status:", final.Status)
}
