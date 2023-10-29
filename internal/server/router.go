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

func PersistenceAPIListenAndServe(router *mux.Router, persistence database.DevicePersistenceDB, attendant adapterattendant.Attendant) error {
	apiv0 := router.PathPrefix("/v0/").Subrouter()

	apiv0.HandleFunc("/devices", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		bridgeKey := reader.Header.Get("Bridge-Key")
		if bridgeKey == "" {
			http.Error(writer, "May not update devices when not identifying as an adapter", http.StatusBadRequest)
			return
		}
		device := devicestoretemplates.Device{}
		err := json.NewDecoder(reader.Body).Decode(&device)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
			return
		}
		// We do not trust the client that much. Override the bridgeKey
		device.BridgeKey = bridgeKey

		_, err = attendant.GetAdapter(bridgeKey, ctx)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}
		err = persistence.UpdateDevice(device, string(bridgeKey), ctx)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}

		writer.WriteHeader(http.StatusOK)

	}).Methods("POST")

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

		adapter, err := attendant.GetAdapter(string(capability.BridgeKey), ctx)
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

	apiv0.HandleFunc("/groups", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		groups, err := persistence.FilterGroups(ctx)
		if err != nil {
			serveHTTPError(err, ctx, writer)
			return
		}

		jsonEncoded, err := json.MarshalIndent(groups, "", "   ")
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}

		writer.Write(jsonEncoded)
	}).Methods("GET")

	apiv0.HandleFunc("/groups", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		bridgeKey := reader.Header.Get("Bridge-Key")
		if bridgeKey == "" {
			serveHTTPError(liberrors.NewApiError(liberrors.UserError, errors.New("Only bridges may update groups")), ctx, writer)
			return
		}
		group := devicestoretemplates.Group{}
		err := json.NewDecoder(reader.Body).Decode(&group)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
			return
		}
		rGroup, err := persistence.UpdateGroup(group, bridgeKey, ctx)
		if err != nil {
			serveHTTPError(err, ctx, writer)
			return
		}
		jsonEncoded, err := json.MarshalIndent(rGroup, "", "   ")
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("POST")

	apiv0.HandleFunc("/groups/{groupID}", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		vars := mux.Vars(reader)
		groupId := vars["groupID"]
		logging.Info(fmt.Sprintf("Getting group with identifier '%s'", groupId), ctx)
		group, err := persistence.GetGroupByIdentifier(groupId, true, ctx)
		if err != nil {
			serveHTTPError(err, ctx, writer)
			return
		}

		jsonEncoded, err := json.MarshalIndent(group, "", "   ")
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("GET")

	apiv0.HandleFunc("/groups/{groupId}/capabilities/{capabilityId}", func(writer http.ResponseWriter, reader *http.Request) {
		ctx := reader.Context()
		vars := mux.Vars(reader)
		groupId := vars["groupId"]
		capabilityId := vars["capabilityId"]
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

		logging.Info(fmt.Sprintf("Triggering capability '%s' of group '%s'", capabilityId, groupId), ctx)
		capability, err := persistence.GetGroupCapability(groupId, capabilityId, ctx)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}

		adapter, err := attendant.GetAdapter(string(capability.CapabilityBridgeKey), ctx)
		if err != nil {
			serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
			return
		}
		sysErr := adapters.TriggerGroupCapability(ctx, adapter, groupId, capabilityId, capArg)
		if err != nil {
			serveHTTPError(sysErr, ctx, writer)
			return
		}
		logging.Info("Capability seemingly successfully triggered", ctx)

	}).Methods("POST")

	return nil

}
