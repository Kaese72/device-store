package server

// apiv0.HandleFunc("/groups", func(writer http.ResponseWriter, reader *http.Request) {
// 	ctx := reader.Context()
// 	groups, err := persistence.FilterGroups(ctx)
// 	if err != nil {
// 		serveHTTPError(err, ctx, writer)
// 		return
// 	}

// 	jsonEncoded, err := json.MarshalIndent(groups, "", "   ")
// 	if err != nil {
// 		serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
// 		return
// 	}

// 	writer.Write(jsonEncoded)
// }).Methods("GET")

// apiv0.HandleFunc("/groups", func(writer http.ResponseWriter, reader *http.Request) {
// 	ctx := reader.Context()
// 	bridgeKey := reader.Header.Get("Bridge-Key")
// 	if bridgeKey == "" {
// 		serveHTTPError(liberrors.NewApiError(liberrors.UserError, errors.New("Only bridges may update groups")), ctx, writer)
// 		return
// 	}
// 	group := devicestoretemplates.Group{}
// 	err := json.NewDecoder(reader.Body).Decode(&group)
// 	if err != nil {
// 		serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
// 		return
// 	}
// 	rGroup, err := persistence.UpdateGroup(group, bridgeKey, ctx)
// 	if err != nil {
// 		serveHTTPError(err, ctx, writer)
// 		return
// 	}
// 	jsonEncoded, err := json.MarshalIndent(rGroup, "", "   ")
// 	if err != nil {
// 		serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
// 		return
// 	}

// 	writer.Write(jsonEncoded)

// }).Methods("POST")

// apiv0.HandleFunc("/groups/{groupID}", func(writer http.ResponseWriter, reader *http.Request) {
// 	ctx := reader.Context()
// 	vars := mux.Vars(reader)
// 	groupId := vars["groupID"]
// 	logging.Info(fmt.Sprintf("Getting group with identifier '%s'", groupId), ctx)
// 	group, err := persistence.GetGroupByIdentifier(groupId, true, ctx)
// 	if err != nil {
// 		serveHTTPError(err, ctx, writer)
// 		return
// 	}

// 	jsonEncoded, err := json.MarshalIndent(group, "", "   ")
// 	if err != nil {
// 		serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
// 		return
// 	}

// 	writer.Write(jsonEncoded)

// }).Methods("GET")

// apiv0.HandleFunc("/groups/{groupId}/capabilities/{capabilityId}", func(writer http.ResponseWriter, reader *http.Request) {
// 	ctx := reader.Context()
// 	vars := mux.Vars(reader)
// 	groupId := vars["groupId"]
// 	capabilityId := vars["capabilityId"]
// 	capArg := devicestoretemplates.CapabilityArgs{}
// 	err := json.NewDecoder(reader.Body).Decode(&capArg)
// 	if err != nil {
// 		if err == io.EOF {
// 			capArg = devicestoretemplates.CapabilityArgs{}

// 		} else {
// 			serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
// 			return
// 		}
// 	}

// 	logging.Info(fmt.Sprintf("Triggering capability '%s' of group '%s'", capabilityId, groupId), ctx)
// 	capability, err := persistence.GetGroupCapability(groupId, capabilityId, ctx)
// 	if err != nil {
// 		serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
// 		return
// 	}

// 	adapter, err := attendant.GetAdapter(string(capability.CapabilityBridgeKey), ctx)
// 	if err != nil {
// 		serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
// 		return
// 	}
// 	sysErr := adapters.TriggerGroupCapability(ctx, adapter, groupId, capabilityId, capArg)
// 	if sysErr != nil {
// 		serveHTTPError(sysErr, ctx, writer)
// 		return
// 	}
// 	logging.Info("Capability seemingly successfully triggered", ctx)

// }).Methods("POST")
