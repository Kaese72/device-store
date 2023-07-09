package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/adapters"
	"github.com/Kaese72/device-store/internal/database"
	"github.com/Kaese72/device-store/internal/logging"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"github.com/Kaese72/huemie-lib/liberrors"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.elastic.co/apm/module/apmgorilla/v2"
)

type apiModelError struct {
	Message string `json:"message"`
}

func serveHTTPError(err error, ctx context.Context, writer http.ResponseWriter) {
	var apiError *liberrors.ApiError
	if !errors.As(err, &apiError) {
		apiError = liberrors.NewApiError(liberrors.InternalError, errors.New("unknown internal error occured"))
	}

	if apiError.Reason <= 599 && apiError.Reason >= 500 {
		// Internal server error, log error
		logging.ErrorErr(err, ctx)
	} else {
		// For everything else, log an info. This may be excessive, but it only happens on errors
		logging.ErrorErr(err, ctx)
	}
	// Safe to ignore error here... kind of but not really
	resp, err := json.Marshal(apiModelError{Message: err.Error()})
	if err != nil {
		logging.ErrorErr(errors.Wrap(err, "failed to marshal error message"), ctx)
		return
	}

	writer.WriteHeader(int(apiError.Reason))
	_, err = writer.Write(resp)
	if err != nil {
		logging.ErrorErr(errors.Wrap(err, "failed to write error message"), ctx)
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

		jsonEncoded, err := json.MarshalIndent(devices, "", "   ")
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
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
			serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
			return
		}
		if len(device.Capabilities) > 0 {
			if bridgeKey == "" {
				http.Error(writer, "May not set capabilities when not identifying as an adapter", http.StatusBadRequest)
				return
			}
			_, err := attendant.GetAdapter(bridgeKey, ctx)
			if err != nil {
				serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
				return
			}
			rDevice, err = persistence.UpdateDeviceAttributesAndCapabilities(device, string(bridgeKey), ctx)
			if err != nil {
				serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
				return
			}
		} else {
			rDevice, err = persistence.UpdateDeviceAttributes(device, true, ctx)
			if err != nil {
				serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
				return
			}
		}

		jsonEncoded, err := json.MarshalIndent(rDevice, "", "   ")
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
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

		jsonEncoded, err := json.MarshalIndent(device, "", "   ")
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
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
				serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
				return
			}

		}

		logging.Info(fmt.Sprintf("Triggering capability '%s' of device '%s'", capabilityID, deviceID), ctx)
		capability, err := persistence.GetCapability(deviceID, capabilityID, ctx)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}

		adapter, err := attendant.GetAdapter(string(capability.CapabilityBridgeKey), ctx)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}
		sysErr := adapters.TriggerDeviceCapability(ctx, adapter, deviceID, capabilityID, capArg)
		if err != nil {
			serveHTTPError(sysErr, ctx, writer)
			return
		}
		logging.Info("Capability seemingly successfully triggered", ctx)

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
