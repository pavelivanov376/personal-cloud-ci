sap.ui.define([
    "sap/ui/core/mvc/Controller",
    "sap/ui/model/json/JSONModel",
    "sap/m/MessageToast",
    "sap/m/MessageBox",
    "sap/ui/core/Item"
], function (Controller, JSONModel, MessageToast, MessageBox, Item) {
    "use strict";

    const API = "/api/jobs";
    const REPO_API = "/api/repositories";

    return Controller.extend("app.controller.Jobs", {

        onInit: function () {
            this.getView().setModel(new JSONModel({ items: [] }), "jobs");
            this.getOwnerComponent().getRouter().getRoute("jobs").attachPatternMatched(this._load, this);
        },

        _load: function () {
            fetch(API)
                .then(r => r.json())
                .then(data => this.getView().getModel("jobs").setProperty("/items", data));
        },

        onAddJob: function () {
            const open = d => {
                const select = this.byId("repositorySelect");
                select.removeAllItems();
                select.addItem(new Item({ key: "", text: "-- Select a repository --" }));
                fetch(REPO_API)
                    .then(r => r.json())
                    .then(repos => repos.forEach(r => select.addItem(new Item({ key: r.id, text: r.name }))));
                d.open();
            };

            if (!this._dialog) {
                this._dialog = this.loadFragment({ name: "app.view.AddJobDialog" });
            }
            this._dialog.then(open);
        },

        onConfirmAdd: function () {
            const nameInput = this.byId("nameInput");
            const select = this.byId("repositorySelect");
            const name = nameInput.getValue().trim();
            const repositoryId = select.getSelectedKey();

            if (!name)         { nameInput.setValueState("Error"); return; }
            if (!repositoryId) { select.setValueState("Error");    return; }
            nameInput.setValueState("None");
            select.setValueState("None");

            fetch(API, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name, repositoryId })
            })
                .then(r => r.json())
                .then(() => { this._closeDialog(); MessageToast.show("Job added"); this._load(); })
                .catch(() => MessageBox.error("Could not add job"));
        },

        onCancelAdd: function () { this._closeDialog(); },

        _closeDialog: function () {
            this.byId("addJobDialog").close();
            this.byId("nameInput").setValue("").setValueState("None");
            this.byId("repositorySelect").setSelectedKey("").setValueState("None");
        },

        onBuild: function (event) {
            const item = event.getSource().getBindingContext("jobs").getObject();
            fetch("/api/builds/jobs/" + item.id, { method: "POST" })
                .then(r => r.json())
                .then(() => { MessageToast.show("Build started"); this._load(); })
                .catch(() => MessageBox.error("Could not start build"));
        },

        onDelete: function (event) {
            const item = event.getSource().getBindingContext("jobs").getObject();
            MessageBox.confirm("Delete job \"" + item.name + "\"?", {
                onClose: action => {
                    if (action !== MessageBox.Action.OK) return;
                    fetch(API + "/" + item.id, { method: "DELETE" })
                        .then(() => { MessageToast.show("Deleted"); this._load(); })
                        .catch(() => MessageBox.error("Could not delete job"));
                }
            });
        }
    });
});
