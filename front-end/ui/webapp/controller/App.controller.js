sap.ui.define([
    "sap/ui/core/mvc/Controller",
    "sap/ui/model/json/JSONModel"
], function (Controller, JSONModel) {
    "use strict";

    return Controller.extend("app.controller.App", {

        onInit: function () {
            this.getView().setModel(new JSONModel({ selectedTab: "repositories" }));
            this.getOwnerComponent().getRouter().attachRouteMatched(this.onRouteMatched, this);
        },

        onRouteMatched: function (event) {
            this.getView().getModel().setProperty("/selectedTab", event.getParameter("name"));
        },

        onTabSelect: function (event) {
            this.getOwnerComponent().getRouter().navTo(event.getParameter("key"));
        }
    });
});
