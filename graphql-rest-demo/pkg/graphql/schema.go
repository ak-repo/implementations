// Package graphql provides GraphQL schema and resolvers.
package graphql

import (
	"graphql-rest-demo/internal/model"
	"graphql-rest-demo/internal/service"

	"github.com/graphql-go/graphql"
)

// Schema holds the GraphQL schema and service.
type Schema struct {
	schema  graphql.Schema
	service *service.BlogService
}

// NewSchema creates a new GraphQL schema.
func NewSchema(service *service.BlogService) (*Schema, error) {
	s := &Schema{service: service}

	// Define Blog type
	blogType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Blog",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"title": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"content": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"author": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"tags": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			},
			"created_at": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// Define Query type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"blogs": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(blogType))),
				Args: graphql.FieldConfigArgument{
					"author": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"tag": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"q": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: s.resolveBlogs,
			},
			"blog": &graphql.Field{
				Type: blogType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: s.resolveBlog,
			},
		},
	})

	// Define Mutation type
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createBlog": &graphql.Field{
				Type: blogType,
				Args: graphql.FieldConfigArgument{
					"title": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"content": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"author": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"tags": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
					},
				},
				Resolve: s.resolveCreateBlog,
			},
		},
	})

	// Create schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		return nil, err
	}

	s.schema = schema
	return s, nil
}

// Execute runs a GraphQL query.
func (s *Schema) Execute(query string, variables map[string]interface{}) *graphql.Result {
	return graphql.Do(graphql.Params{
		Schema:         s.schema,
		RequestString:  query,
		VariableValues: variables,
	})
}

// Resolver functions

func (s *Schema) resolveBlogs(p graphql.ResolveParams) (interface{}, error) {
	filter := model.BlogFilter{
		Author: getStringArg(p.Args, "author"),
		Tag:    getStringArg(p.Args, "tag"),
		Query:  getStringArg(p.Args, "q"),
	}

	return s.service.ListBlogs(filter), nil
}

func (s *Schema) resolveBlog(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)
	return s.service.GetBlog(id)
}

func (s *Schema) resolveCreateBlog(p graphql.ResolveParams) (interface{}, error) {
	req := model.CreateBlogRequest{
		Title:   p.Args["title"].(string),
		Content: p.Args["content"].(string),
		Author:  p.Args["author"].(string),
	}

	// Handle tags
	if tagsArg, ok := p.Args["tags"]; ok && tagsArg != nil {
		if tags, ok := tagsArg.([]interface{}); ok {
			for _, tag := range tags {
				if s, ok := tag.(string); ok {
					req.Tags = append(req.Tags, s)
				}
			}
		}
	}

	return s.service.CreateBlog(req)
}

func getStringArg(args map[string]interface{}, key string) string {
	if val, ok := args[key]; ok && val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
