package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/database"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/systemerrors"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"github.com/gorilla/mux"
	"go.elastic.co/apm/module/apmgorilla/v2"
)

type apiModelError struct {
	Message string `json:"message"`
}

func serveHTTPError(err systemerrors.SystemError, ctx context.Context, writer http.ResponseWriter) {
	if err.Reason() <= 599 && err.Reason() >= 500 {
		// Internal server error, log error
		logging.Error(err.Error(), ctx)
	} else {
		// For everything else, log an info. This may be excessive, but it only happens on errors
		logging.Info(err.Error(), ctx)
	}
	// Safe to ignore error here... kind of but not really
	resp, err2 := json.Marshal(apiModelError{Message: err.Error()})
	if err2 != nil {
		logging.Error(err2.Error(), ctx)
		return
	}

	writer.WriteHeader(err.Reason())
	_, err2 = writer.Write(resp)
	if err2 != nil {
		logging.Error(err2.Error(), ctx)
	}
}

func PersistenceAPIListenAndServe(persistence database.DevicePersistenceDB, attendant adapterattendant.Attendant) error {
	router := mux.NewRouter()
	apmgorilla.Instrument(router)

	//Everything else (not /auth/login) should have the authentication middleware
	apiv0 := router.PathPrefix("/device-store/v0/").Subrouter()

	apiv0.HandleFunc("/devices", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		devices, err := persistence.FilterDevices(ctx)
		if err != nil {
			serveHTTPError(err, ctx, writer)
			return
		}

		jsonEncoded, err2 := json.MarshalIndent(devices, "", "   ")
		if err2 != nil {
			serveHTTPError(systemerrors.WrapSystemError(err2, systemerrors.InternalError), ctx, writer)
			return
		}

		writer.Write(jsonEncoded)
	}).Methods("GET")

	apiv0.HandleFunc("/devices", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		bridgeKey := reader.Header.Get("Bridge-Key")
		device := devicestoretemplates.Device{}
		var rDevice devicestoretemplates.Device
		err := json.NewDecoder(reader.Body).Decode(&device)
		if err != nil {
			serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.UserError), ctx, writer)
			return
		}
		if len(device.Capabilities) > 0 {
			if bridgeKey == "" {
				http.Error(writer, "May not set capabilities when not identifying as an adapter", http.StatusBadRequest)
				return
			}
			_, err := attendant.GetAdapter(bridgeKey)
			if err != nil {
				serveHTTPError(systemerrors.WrapSystemError(fmt.Errorf("could not get adapter, '%s'", err.Error()), systemerrors.NotFound), ctx, writer)
				return
			}
			rDevice, err = persistence.UpdateDeviceAttributesAndCapabilities(device, string(bridgeKey), ctx)
			if err != nil {
				serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
				return
			}
		} else {
			rDevice, err = persistence.UpdateDeviceAttributes(device, true, ctx)
			if err != nil {
				serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
				return
			}
		}

		jsonEncoded, err := json.MarshalIndent(rDevice, "", "   ")
		if err != nil {
			serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("POST")

	apiv0.HandleFunc("/devices/{deviceID}", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		vars := mux.Vars(reader)
		deviceID := vars["deviceID"]
		logging.Info(fmt.Sprintf("Getting device with identifier '%s'", deviceID), ctx)
		device, err := persistence.GetDeviceByIdentifier(deviceID, true, ctx)
		if err != nil {
			serveHTTPError(err, ctx, writer)
			return
		}

		jsonEncoded, err2 := json.MarshalIndent(device, "", "   ")
		if err2 != nil {
			serveHTTPError(systemerrors.WrapSystemError(err2, systemerrors.InternalError), ctx, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("GET")

	apiv0.HandleFunc("/devices/{deviceID}/capabilities/{capabilityID}", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		vars := mux.Vars(reader)
		deviceID := vars["deviceID"]
		capabilityID := vars["capabilityID"]
		capArg := devicestoretemplates.CapabilityArgs{}
		err := json.NewDecoder(reader.Body).Decode(&capArg)
		if err != nil {
			if err == io.EOF {
				capArg = devicestoretemplates.CapabilityArgs{}

			} else {
				serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.UserError), ctx, writer)
				return
			}

		}

		logging.Info(fmt.Sprintf("Triggering capability '%s' of device '%s'", capabilityID, deviceID), ctx)
		//err = persistence.TriggerCapability(deviceID, capabilityID, capArg)
		capability, err := persistence.GetCapability(deviceID, capabilityID, ctx)
		if err != nil {
			serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
			return
		}
		jsonEncoded, err := json.Marshal(capArg)
		if err != nil {
			serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
			return
		}

		adapter, err := attendant.GetAdapter(string(capability.CapabilityBridgeKey))
		if err != nil {
			serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
			return
		}
		adapterURL, err := url.Parse(adapter.Address)
		if err != nil {
			serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
			return
		}

		adapterURL.Path = fmt.Sprintf("devices/%s/capabilities/%s", deviceID, capabilityID)
		logging.Info("Triggering capability", ctx, map[string]interface{}{"capUri": adapterURL.String()})
		resp, err := http.Post(adapterURL.String(), "application/json", bytes.NewBuffer(jsonEncoded))
		if err != nil {
			// FIXME What if there is interesting debug information in the response?
			// We should log it or incorporate it in the response message or something
			serveHTTPError(systemerrors.WrapSystemError(err, systemerrors.InternalError), ctx, writer)
			return
		}
		// It is the callers responsibility to Close the body reader
		// But there should not be anything of interest here at the moment
		defer resp.Body.Close()
		logging.Info("Capability triggered", ctx, map[string]interface{}{"rCode": strconv.Itoa(resp.StatusCode)})

	}).Methods("POST")

	server := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8080",
	}

	if err := server.ListenAndServe(); err != nil {
		logging.Error(err.Error(), context.TODO())
		return err
	}

	return nil

}
