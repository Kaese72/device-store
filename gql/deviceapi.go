package gql

import (
	"context"
	"errors"
	"net/http"

	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/rest/models"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type gqlpersistenceinterface interface {
	GetDeviceByIdentifier(string, bool, context.Context) (models.Device, error)
	FilterDevices(context.Context) ([]models.Device, error)
}

type gDevice struct {
	Identifier string `json:"identifier"`
}

func deviceFromIntermediary(mDevice models.Device) gDevice {
	return gDevice{
		Identifier: mDevice.Identifier,
	}
}

func GraphQLListenAndServe(persistence gqlpersistenceinterface) error {
	device := graphql.NewObject(graphql.ObjectConfig{
		Name: "Device",
		Fields: graphql.Fields{
			"identifier": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// attribute := graphql.NewObject(graphql.ObjectConfig{
	// 	Name: "Attribute",
	// 	Fields: graphql.Fields{
	// 		"booleanState": &graphql.Field{
	// 			Type: graphql.Boolean,
	// 		},
	// 		"numericState": &graphql.Field{
	// 			Type: graphql.Int,
	// 		},
	// 		"stringState": &graphql.Field{
	// 			Type: graphql.String,
	// 		},
	// 	},
	// })

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"device": &graphql.Field{
				Type:        device,
				Description: "A single device",
				Args: graphql.FieldConfigArgument{
					"identifier": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					identifierQuery, ok := params.Args["identifier"].(string)
					if !ok {
						return nil, errors.New("must supply identifier query parameter")
					}
					mDevice, err := persistence.GetDeviceByIdentifier(identifierQuery, false, context.TODO())
					if err != nil {
						logging.Error(err.Error(), context.TODO())
						return nil, err
					}
					return deviceFromIntermediary(mDevice), nil
				},
			},
			"deviceList": &graphql.Field{
				Type:        device,
				Description: "List of devices",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					mDevices, err := persistence.FilterDevices(context.TODO())
					if err != nil {
						logging.Error(err.Error(), context.TODO())
						return nil, err
					}
					var rDevices []gDevice
					for _, mDevice := range mDevices {
						rDevices = append(rDevices, deviceFromIntermediary(mDevice))
					}
					return rDevices, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	if err != nil {
		panic(err)
	}
	server := &http.Server{
		Handler: handler.New(&handler.Config{
			Schema:   &schema,
			Pretty:   true,
			GraphiQL: false,
		}),
		Addr: "0.0.0.0:8081",
	}
	err = server.ListenAndServe()
	if err != nil {
		logging.Error(err.Error(), context.TODO())
	}
	return err
}
