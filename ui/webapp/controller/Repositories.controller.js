sap.ui.define([
    "sap/ui/core/mvc/Controller",
    "sap/ui/model/json/JSONModel",
    "sap/m/MessageToast",
    "sap/m/MessageBox"
], function (Controller, JSONModel, MessageToast, MessageBox) {
    "use strict";

    const API = "/api/repositories";

    return Controller.extend("app.controller.Repositories", {

        onInit: function () {
            this.getView().setModel(new JSONModel({ items: [] }), "repositories");
            this._load();
        },

        _load: function () {
            fetch(API)
                .then(r => r.json())
                .then(data => this.getView().getModel("repositories").setProperty("/items", data));
        },

        onAddRepository: function () {
            if (!this._dialog) {
                this._dialog = this.loadFragment({ name: "app.view.AddRepositoryDialog" });
            }
            this._dialog.then(d => d.open());
        },

        onConfirmAdd: function () {
            const nameInput = this.byId("nameInput");
            const urlInput = this.byId("urlInput");
            const name = nameInput.getValue().trim();
            const url = urlInput.getValue().trim();

            if (!name) { nameInput.setValueState("Error"); return; }
            if (!url)  { urlInput.setValueState("Error");  return; }
            nameInput.setValueState("None");
            urlInput.setValueState("None");

            fetch(API, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name, url })
            })
                .then(r => r.json())
                .then(() => { this._closeDialog(); MessageToast.show("Repository added"); this._load(); })
                .catch(() => MessageBox.error("Could not add repository"));
        },

        onCancelAdd: function () { this._closeDialog(); },

        _closeDialog: function () {
            this.byId("addDialog").close();
            this.byId("nameInput").setValue("").setValueState("None");
            this.byId("urlInput").setValue("").setValueState("None");
        },

        onDelete: function (event) {
            const item = event.getSource().getBindingContext("repositories").getObject();
            MessageBox.confirm("Delete repository \"" + item.name + "\"?", {
                onClose: action => {
                    if (action !== MessageBox.Action.OK) return;
                    fetch(API + "/" + item.id, { method: "DELETE" })
                        .then(() => { MessageToast.show("Deleted"); this._load(); })
                        .catch(() => MessageBox.error("Could not delete repository"));
                }
            });
        }
    });
});
