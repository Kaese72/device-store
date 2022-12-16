package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kaese72/device-store/adapterattendant"
	"github.com/Kaese72/device-store/config"
	"github.com/Kaese72/device-store/database"
	"github.com/Kaese72/sdup-lib/devicestoretemplates"
	"github.com/Kaese72/sdup-lib/logging"
	"github.com/gorilla/mux"
)

func ServeHTTPError(err error, writer http.ResponseWriter) {
	switch err.(type) {
	case database.NotFound:
		http.Error(writer, "Not Found", http.StatusNotFound)
	case database.UserError:
		http.Error(writer, "Bad Request", http.StatusBadRequest)
	default:
		http.Error(writer, "Internal Server", http.StatusInternalServerError)
	}
}

func PersistenceAPIListenAndServe(config config.HTTPConfig, persistence database.DevicePersistenceDB, attendant adapterattendant.Attendant) error {
	router := mux.NewRouter()

	//Everything else (not /auth/login) should have the authentication middleware
	apiv0 := router.PathPrefix("/rest/v0/").Subrouter()

	apiv0.HandleFunc("/devices", func(writer http.ResponseWriter, reader *http.Request) {
		devices, err := persistence.FilterDevices()
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		jsonEncoded, err := json.MarshalIndent(devices, "", "   ")
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)
	}).Methods("GET")

	apiv0.HandleFunc("/devices", func(writer http.ResponseWriter, reader *http.Request) {
		bridgeKey := reader.Header.Get("Bridge-Key")
		device := devicestoretemplates.Device{}
		var rDevice devicestoretemplates.Device
		err := json.NewDecoder(reader.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if len(device.Capabilities) > 0 {
			if bridgeKey == "" {
				http.Error(writer, "May not set capabilities when not identifying as an adapter", http.StatusBadRequest)
				return
			}
			_, err := attendant.GetAdapter(bridgeKey)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Could not get adapter, '%s'", err.Error()), http.StatusBadRequest)
				return
			}
			rDevice, err = persistence.UpdateDeviceAttributesAndCapabilities(device, devicestoretemplates.BridgeKey(bridgeKey))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			rDevice, err = persistence.UpdateDeviceAttributes(device, true)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		jsonEncoded, err := json.MarshalIndent(rDevice, "", "   ")
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("POST")

	apiv0.HandleFunc("/devices/{deviceID}", func(writer http.ResponseWriter, reader *http.Request) {
		vars := mux.Vars(reader)
		deviceID := vars["deviceID"]
		logging.Info(fmt.Sprintf("Getting device with identifier '%s'", deviceID))
		device, err := persistence.GetDeviceByIdentifier(deviceID, true)
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		jsonEncoded, err := json.MarshalIndent(device, "", "   ")
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("GET")

	apiv0.HandleFunc("/devices/{deviceID}/capabilities/{capabilityID}", func(writer http.ResponseWriter, reader *http.Request) {
		vars := mux.Vars(reader)
		deviceID := vars["deviceID"]
		capabilityID := vars["capabilityID"]
		capArg := devicestoretemplates.CapabilityArgs{}
		err := json.NewDecoder(reader.Body).Decode(&capArg)
		if err != nil {
			if err == io.EOF {
				capArg = devicestoretemplates.CapabilityArgs{}

			} else {
				ServeHTTPError(database.UserError(err), writer)
				return
			}

		}
		logging.Info(fmt.Sprintf("Triggering capability '%s' of device '%s'", capabilityID, deviceID))
		err = persistence.TriggerCapability(deviceID, capabilityID, capArg)
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

	}).Methods("POST")

	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf("%s:%d", config.Address, config.Port),
	}

	if err := server.ListenAndServe(); err != nil {
		logging.Error(err.Error())
		return err
	}

	return nil

}
