package gql

import (
	"context"
	"errors"
	"time"

	"github.com/Kaese72/device-store/internal/database"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/models/intermediaries"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type gDevice struct {
	BridgeIdentifier string `json:"bridgeIdentifier"`
	BridgeKey        string `json:"bridgeKey"`
	StoreIdentifier  string `json:"storeIdentifier"`
}

func deviceFromIntermediary(mDevice intermediaries.DeviceIntermediary) gDevice {
	return gDevice{
		BridgeIdentifier: mDevice.BridgeIdentifier,
		BridgeKey:        mDevice.BridgeKey,
		StoreIdentifier:  mDevice.DeviceStoreIdentifier,
	}
}

type gAttribute struct {
	Name         string   `json:"name"`
	BooleanState *bool    `json:"booleanState,omitempty"`
	TextState    *string  `json:"textState,omitempty"`
	NumericState *float32 `json:"numericState,omitempty"`
}

func attributeFromIntermediary(mAttribute intermediaries.AttributeIntermediary) gAttribute {
	return gAttribute{
		Name:         mAttribute.Name,
		BooleanState: mAttribute.State.Boolean,
		TextState:    mAttribute.State.Text,
		NumericState: mAttribute.State.Numeric,
	}
}

type gCapability struct {
	Name      string    `json:"name"`
	BridgeKey string    `json:"bridgeKey"`
	LastSeen  time.Time `json:"lastSeen"`
}

func capabilityFromIntermediary(mCapability intermediaries.CapabilityIntermediary) gCapability {
	return gCapability{
		Name:      mCapability.Name,
		BridgeKey: mCapability.BridgeKey,
		LastSeen:  mCapability.LastSeen,
	}
}

func GraphQLListenAndServe(router *mux.Router, persistence database.DevicePersistenceDB) error {
	attribute := graphql.NewObject(graphql.ObjectConfig{
		Name: "Attribute",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"booleanState": &graphql.Field{
				Type: graphql.Boolean,
			},
			"textState": &graphql.Field{
				Type: graphql.String,
			},
			"numericState": &graphql.Field{
				Type: graphql.Float,
			},
		},
	})

	capability := graphql.NewObject(graphql.ObjectConfig{
		Name: "Capability",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"bridgeKey": &graphql.Field{
				Type: graphql.String,
			},
			"lastSeen": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	})

	device := graphql.NewObject(graphql.ObjectConfig{
		Name: "Device",
		Fields: graphql.Fields{
			"bridgeIdentifier": &graphql.Field{
				Type: graphql.String,
			},
			"bridgeKey": &graphql.Field{
				Type: graphql.String,
			},
			"storeIdentifier": &graphql.Field{
				Type: graphql.String,
			},
			"attributes": &graphql.Field{
				Type:        graphql.NewList(attribute),
				Description: "Attributes belonging to a device",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					sDevice, ok := params.Source.(gDevice)
					if !ok {
						return nil, errors.New("could not translate source")
					}
					iAttributes, err := persistence.GetDeviceAttributes(sDevice.StoreIdentifier, context.TODO())
					if err != nil {
						return nil, err
					}
					var rAttributes []gAttribute
					for _, iAttribute := range iAttributes {
						rAttributes = append(rAttributes, attributeFromIntermediary(iAttribute))
					}
					return rAttributes, nil
				},
			},
			"capabilities": &graphql.Field{
				Type:        graphql.NewList(capability),
				Description: "Capabilities belonging to a device",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					sDevice, ok := params.Source.(gDevice)
					if !ok {
						return nil, errors.New("could not translate source")
					}
					iCapabilities, err := persistence.GetDeviceCapabilities(sDevice.StoreIdentifier, context.TODO())
					if err != nil {
						return nil, err
					}
					var rCapabilities []gCapability
					for _, iCapability := range iCapabilities {
						rCapabilities = append(rCapabilities, capabilityFromIntermediary(iCapability))
					}
					return rCapabilities, nil
				},
			},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"device": &graphql.Field{
				Type:        device,
				Description: "A single device",
				Args: graphql.FieldConfigArgument{
					"storeIdentifier": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					storeIdentifier, ok := params.Args["storeIdentifier"].(string)
					if !ok {
						return nil, errors.New("must supply identifier query parameter")
					}
					mDevice, err := persistence.GetStoreDevice(storeIdentifier, false, context.TODO())
					if err != nil {
						logging.Error(err.Error(), context.TODO())
						return nil, err
					}
					return deviceFromIntermediary(mDevice), nil
				},
			},
			"deviceList": &graphql.Field{
				Type:        graphql.NewList(device),
				Description: "List of devices",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					mDevices, err := persistence.GetDevices(context.TODO())
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
		return err
	}

	router.Handle("/", handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	}))
	return nil
}
