package mock

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"go-slack-ics/system"
)

type ShopifyGraphQl struct {
	redis  *system.Redis
	schema graphql.Schema
	// store wird pro Request gesetzt (Achtung: Bei parallelen Requests kann dies zu Datenrennen führen – in der Produktion besser per Context arbeiten)
	store string
}

func NewShopifyGraphQl() *ShopifyGraphQl {
	s := &ShopifyGraphQl{
		redis: system.NewRedis(),
	}

	// Definiere den Typ "AppSubscription"
	appSubscriptionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AppSubscription",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.ID},
			"name":      &graphql.Field{Type: graphql.String},
			"status":    &graphql.Field{Type: graphql.String},
			"createdAt": &graphql.Field{Type: graphql.String},
			"updatedAt": &graphql.Field{Type: graphql.String},
			"currency":  &graphql.Field{Type: graphql.String},
			"returnUrl": &graphql.Field{Type: graphql.String},
			"test":      &graphql.Field{Type: graphql.Boolean},
		},
	})

	// Definiere den Typ "AppInstallation"
	appInstallationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AppInstallation",
		Fields: graphql.Fields{
			"id": &graphql.Field{Type: graphql.ID},
			"activeSubscriptions": &graphql.Field{
				Type: graphql.NewList(appSubscriptionType),
			},
		},
	})

	// Query: Liefert die aktuelle AppInstallation für den aktuellen Store aus Redis.
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"currentAppInstallation": &graphql.Field{
				Type: appInstallationType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// Den Store aus unserer ShopifyGraphQl-Instanz verwenden
					store := s.store
					key := "appInstallation:" + store

					// Versuche, ein Abonnement für diesen Store aus Redis zu laden
					var subscription map[string]interface{}
					err := s.redis.Get(key, &subscription)
					if err != nil {
						// Kein Eintrag gefunden, also leere Installation zurückgeben
						return map[string]interface{}{
							"id":                  "app_installation_" + store,
							"activeSubscriptions": []interface{}{},
						}, nil
					}

					// Ein Eintrag gefunden – gib ihn als Liste zurück
					return map[string]interface{}{
						"id":                  "app_installation_" + store,
						"activeSubscriptions": []interface{}{subscription},
					}, nil
				},
			},
		},
	})

	// Mutation: Erzeugt ein neues App-Subscription und speichert es in Redis für den aktuellen Store.
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"appSubscriptionCreate": &graphql.Field{
				Type: appSubscriptionType,
				Args: graphql.FieldConfigArgument{
					"name":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"returnUrl": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"test":      &graphql.ArgumentConfig{Type: graphql.Boolean},
					"trialDays": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// Erzeuge eine neue Subscription
					id := strconv.Itoa(int(time.Now().Unix()))
					subscription := map[string]interface{}{
						"id":        id,
						"name":      p.Args["name"].(string),
						"status":    "PENDING",
						"createdAt": time.Now().Format(time.RFC3339),
						"updatedAt": time.Now().Format(time.RFC3339),
						"returnUrl": p.Args["returnUrl"].(string),
						"test":      p.Args["test"].(bool),
						"trialDays": p.Args["trialDays"].(int),
						// Wünschenswert wären hier evtl. auch "currency" etc.
					}

					// Speichere das neue Subscription in Redis unter dem Key für den aktuellen Store.
					key := "appInstallation:" + s.store
					// Überschreibe einen eventuell vorhandenen Eintrag.
					s.redis.Set(key, subscription)

					return subscription, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
	s.schema = schema

	return s
}

func (s *ShopifyGraphQl) GraphQLHandler(c *gin.Context) {
	// Lese den "store"-Parameter aus der URL (z.B. /admin/:store/api/:version/graphql)
	s.store = c.Param("store")

	h := handler.New(&handler.Config{
		Schema:   &s.schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Den Store (falls benötigt) in den Context einfügen – hier erfolgt die Übergabe bereits über s.store
	ctx := context.WithValue(c.Request.Context(), "shopify", s)
	c.Request = c.Request.WithContext(ctx)
	h.ServeHTTP(c.Writer, c.Request)
}
