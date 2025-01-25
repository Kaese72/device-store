package restwebapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/adapters"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence"
	"github.com/Kaese72/device-store/restmodels"
	"github.com/Kaese72/huemie-lib/liberrors"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type webApp struct {
	persistence persistence.RestPersistenceDB
	attendant   adapterattendant.Attendant
}

func NewWebApp(persistence persistence.RestPersistenceDB, attendant adapterattendant.Attendant) webApp {
	return webApp{
		persistence: persistence,
		attendant:   attendant,
	}
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
	resp, err := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: err.Error()})
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

// GetDevices returns all devices in the database
func (app webApp) GetDevices(writer http.ResponseWriter, reader *http.Request) {
	ctx := reader.Context()
	restDevices, err := app.persistence.GetDevices(ctx, restmodels.ParseQueryIntoFilters(reader.URL.Query()))
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}
	resp, err := json.Marshal(restDevices)
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(resp)
	if err != nil {
		serveHTTPError(err, ctx, writer)
	}
}

func (app webApp) TriggerDeviceCapability(writer http.ResponseWriter, reader *http.Request) {
	ctx := reader.Context()
	vars := mux.Vars(reader)
	// Because of regex above this will never happen
	storeDeviceIdentifier, _ := strconv.Atoi(vars["storeDeviceIdentifier"])
	capabilityID := vars["capabilityID"]
	capArg := restmodels.DeviceCapabilityArgs{}
	err := json.NewDecoder(reader.Body).Decode(&capArg)
	if err != nil {
		if err == io.EOF {
			capArg = restmodels.DeviceCapabilityArgs{}

		} else {
			serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
			return
		}

	}

	logging.Info(fmt.Sprintf("Triggering capability '%s' of device '%d'", capabilityID, storeDeviceIdentifier), ctx)
	capability, err := app.persistence.GetDeviceCapabilityForActivation(ctx, storeDeviceIdentifier, capabilityID)
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}

	adapter, err := app.attendant.GetAdapter(string(capability.BridgeKey), ctx)
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}
	sysErr := adapters.TriggerDeviceCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, capArg)
	if sysErr != nil {
		serveHTTPError(sysErr, ctx, writer)
		return
	}
	logging.Info("Capability seemingly successfully triggered", ctx)
}

func (app webApp) GetGroups(writer http.ResponseWriter, reader *http.Request) {
	ctx := reader.Context()
	restGroups, err := app.persistence.GetGroups(ctx, restmodels.ParseQueryIntoFilters(reader.URL.Query()))
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}
	resp, err := json.Marshal(restGroups)
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(resp)
	if err != nil {
		serveHTTPError(err, ctx, writer)
	}
}

func (app webApp) TriggerGroupCapability(writer http.ResponseWriter, reader *http.Request) {
	ctx := reader.Context()
	vars := mux.Vars(reader)
	storeGroupIdentifier, _ := strconv.Atoi(vars["storeGroupIdentifier"])
	capabilityID := vars["capabilityID"]
	capArg := restmodels.DeviceCapabilityArgs{}
	err := json.NewDecoder(reader.Body).Decode(&capArg)
	if err != nil {
		if err == io.EOF {
			capArg = restmodels.DeviceCapabilityArgs{}

		} else {
			serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
			return
		}

	}

	logging.Info(fmt.Sprintf("Triggering capability '%s' of group '%d'", capabilityID, storeGroupIdentifier), ctx)
	capability, err := app.persistence.GetGroupCapabilityForActivation(ctx, storeGroupIdentifier, capabilityID)
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}

	adapter, err := app.attendant.GetAdapter(string(capability.BridgeKey), ctx)
	if err != nil {
		serveHTTPError(err, ctx, writer)
		return
	}
	sysErr := adapters.TriggerGroupCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, capArg)
	if sysErr != nil {
		serveHTTPError(sysErr, ctx, writer)
		return
	}
	logging.Info("Capability seemingly successfully triggered", ctx)
}
