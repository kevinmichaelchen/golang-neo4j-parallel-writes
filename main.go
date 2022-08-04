package main

import (
	"fmt"
	"github.com/jaswdr/faker"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	uri      = `bolt://localhost:7687`
	username = "neo4j"
	password = "neo4j"

	minNumberActionsPerUser = 3
	maxNumberActionsPerUser = 10

	numUsers   = 10
	timeFormat = "2006-01-02T15:04:05.999-0700"
)

func main() {
	// Set pretty logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create driver
	driver := newDriver()
	defer driver.Close()

	// Create session
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	// Create union
	createUnionAndCampaign(session)

	// Create a ton of people
	names := createPeople(session)

	createRandomActions(session, names)
}

func newDriver() neo4j.Driver {
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatal().Err(err)
	}
	return driver
}

const createUnionAndCampaignQuery = `
CREATE (u:Union {name:"United Steelworkers"}) - [:HAS_CAMPAIGN] -> (c:Campaign {id:1})
`

func createUnionAndCampaign(session neo4j.Session) {
	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			createUnionAndCampaignQuery,
			map[string]interface{}{})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})

	if err != nil {
		log.Fatal().Err(err)
	}
}

func createPeople(session neo4j.Session) []string {
	// We'll put a bunch of people in the Campaign
	var names []string
	f := faker.New()
	for i := 0; i < numUsers; i++ {
		name := f.Person().FirstName() + f.Person().LastName()
		name = strings.ToLower(name)
		name = strings.ReplaceAll(name, "\"", "")
		name = strings.ReplaceAll(name, " ", "-")
		names = append(names, name+xid.New().String())
	}

	// Using string builder for efficient concatenation
	var b strings.Builder
	fmt.Fprint(&b, `MATCH (c:Campaign) WHERE c.id = 1 CREATE `)
	for _, name := range names {
		fmt.Fprintf(&b, `(c)-[:HAS_USER]->(:User {name:"%s"}), `, name)
	}
	query := b.String()       // no copying
	query = query[:b.Len()-2] // no copying (removes trailing ", ")

	log.Info().Msg(query)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			query,
			map[string]interface{}{})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})

	if err != nil {
		log.Fatal().Err(err)
	}

	return names
}

func createRandomActions(session neo4j.Session, names []string) {
	// Generate times
	var times []time.Time
	for _, name := range names {
		numActions := minNumberActionsPerUser + rand.Intn(maxNumberActionsPerUser-minNumberActionsPerUser+1)
		for i := 0; i < numActions; i++ {
			t := time.Now().Add(-1 * time.Hour).
				Add(time.Duration(rand.Intn(60)) * time.Minute).
				Add(time.Duration(rand.Intn(60)) * time.Second).
				Add(time.Duration(rand.Intn(1000)) * time.Millisecond)
			times = append(times, t)
		}
		// Persist them
		createActions(session, name, times)
	}
}

func createActions(session neo4j.Session, name string, times []time.Time) {
	// Using string builder for efficient concatenation
	var b strings.Builder
	fmt.Fprint(&b, `MATCH (u:User) WHERE u.name = $name CREATE `)
	for _, t := range times {
		fmt.Fprintf(&b, `(u)-[:PERFORMED]->(:Action {name: "ACTION_FOO", time: datetime('%s')}), `, t.Format(timeFormat))
	}
	query := b.String()       // no copying
	query = query[:b.Len()-2] // no copying (removes trailing ", ")

	log.Info().Msg(query)

	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			query,
			map[string]interface{}{"name": name})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})

	if err != nil {
		log.Fatal().Err(err)
	}
}
