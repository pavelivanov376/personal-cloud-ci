sap.ui.define(["sap/ui/core/mvc/Controller", "sap/ui/model/json/JSONModel"], function (Controller, JSONModel) {
    return Controller.extend("app.controller.Repositories", {
        onInit: function () {
            fetch("/api/repositories")
                .then(r => r.json())
                .then(data => this.getView().setModel(new JSONModel({ repositories: data })));
        }
    });
});
