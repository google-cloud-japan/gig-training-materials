// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"

	"cloud.google.com/go/spanner"
	"golang.org/x/net/context"
)

type Players struct {
	PlayerId string `spanner:"playerId" json:"playerId"`
	Name     string `spanner:"name" json:"name"`
	Level    int64  `spanner:"level" json:"level"`
	Money    int64  `spanner:"money" json:"money"`
}

const dbName = "player-db"

func NewPlayers() *Players {
	return &Players{}
}

// func GetSpannerInstanceFromEnv returns the name of spanner instance and database, using env ${GOOGLE_CLOUD_PROJECT}
func GetSpannerInstanceFromEnv() string {
	si := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if si == "" {
		log.Fatal("'GOOGLE_CLOUD_PROJECT' is empty. Set 'GOOGLE_CLOUD_PROJECT' env by 'export GOOGLE_CLOUD_PROJECT=<gcp project id>'")
	}
	// return instance and dbname: 'projects/${GOOGLE_CLOUD_PROJECT}/instances/dev-instance/databases/player-db'
	return "projects/" + si + "/instances/dev-instance/databases/" + dbName
}

// func CreateClient returns the client of Cloud Spanner
func CreateClient(ctx context.Context, db string) *spanner.Client {
	client, err := spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
