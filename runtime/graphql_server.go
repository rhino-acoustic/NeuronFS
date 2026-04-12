package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// ============================================================================
// Module: GraphQL API System (Phase 32)
// Provides a structured B2B interface to query the NeuronFS brain state.
// ============================================================================

// buildGraphQLSchema creates the schema for Brain and Neurons
func buildGraphQLSchema(brainRoot string) (graphql.Schema, error) {
	// Object Type: Neuron
	neuronType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Neuron",
			Fields: graphql.Fields{
				"path": &graphql.Field{Type: graphql.String},
				"name": &graphql.Field{Type: graphql.String},
			},
		},
	)

	// Object Type: Region
	regionType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Region",
			Fields: graphql.Fields{
				"name":    &graphql.Field{Type: graphql.String},
				"neurons": &graphql.Field{Type: graphql.NewList(neuronType)},
				"count":   &graphql.Field{Type: graphql.Int},
			},
		},
	)

	// Object Type: TemporalDelta (Phase 38)
	temporalDeltaType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "TemporalDelta",
			Fields: graphql.Fields{
				"timestamp": &graphql.Field{Type: graphql.String}, // Stringified unix ms
				"hash":      &graphql.Field{Type: graphql.String},
				"content":   &graphql.Field{Type: graphql.String},
			},
		},
	)

	// Root Query
	queryType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"brain": &graphql.Field{
					Type: graphql.NewList(regionType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						var result []map[string]interface{}
						
						brain := scanBrain(brainRoot)
						for _, reg := range brain.Regions {
							var neurons []map[string]interface{}
							for _, n := range reg.Neurons {
								neurons = append(neurons, map[string]interface{}{
									"path": n.Path,
									"name": n.Name,
								})
							}
							result = append(result, map[string]interface{}{
								"name":    reg.Name,
								"neurons": neurons,
								"count":   len(reg.Neurons),
							})
						}
						return result, nil
					},
				},
				"status": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return "STABLE / GRAPHQL-ENABLED", nil
					},
				},
				"temporalHistory": &graphql.Field{
					Type: graphql.NewList(temporalDeltaType),
					Args: graphql.FieldConfigArgument{
						"filename": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						filename := p.Args["filename"].(string)
						temporalDir := filepath.Join(brainRoot, "hippocampus", "temporal_log")
						
						var deltas []map[string]interface{}
						
						_ = filepath.WalkDir(temporalDir, func(path string, d fs.DirEntry, err error) error {
							if err != nil || d.IsDir() {
								return nil
							}
							
							// Format: [timestamp]_[filename]_[hash].delta
							base := d.Name()
							if strings.HasSuffix(base, ".delta") && strings.Contains(base, "_"+filename+"_") {
								parts := strings.Split(base, "_")
								if len(parts) >= 3 {
									timestamp := parts[0]
									hashPart := parts[len(parts)-1]
									hashPart = strings.TrimSuffix(hashPart, ".delta")
									
									contentBytes, _ := os.ReadFile(path)
									
									deltas = append(deltas, map[string]interface{}{
										"timestamp": timestamp,
										"hash":      hashPart,
										"content":   string(contentBytes),
									})
								}
							}
							return nil
						})
						
						// Sort by timestamp descending
						sort.Slice(deltas, func(i, j int) bool {
							tsI, _ := strconv.ParseInt(deltas[i]["timestamp"].(string), 10, 64)
							tsJ, _ := strconv.ParseInt(deltas[j]["timestamp"].(string), 10, 64)
							return tsI > tsJ
						})
						
						return deltas, nil
					},
				},
			},
		},
	)

	// Root Mutation
	mutationType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"triggerWebhook": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"event": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
						"message": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						event := p.Args["event"].(string)
						message := p.Args["message"].(string)
						TriggerWebhook(event, message, nil)
						return fmt.Sprintf("Triggered %s", event), nil
					},
				},
			},
		},
	)

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

// HandleGraphQL returns the standard HTTP handler for /graphql Endpoint
func HandleGraphQL(brainRoot string) http.HandlerFunc {
	schema, err := buildGraphQLSchema(brainRoot)
	if err != nil {
		fmt.Printf("[GraphQL] Failed to initialize schema: %v\n", err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "GraphQL Schema Error", 500)
		}
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
