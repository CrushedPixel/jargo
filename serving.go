package jargo

import (
	"github.com/crushedpixel/ferry"
	"github.com/crushedpixel/http_bridge"
	"net/http"
)

// BridgeRoot registers all of the application's controller's actions
// with a Ferry instance at root level.
func (app *Application) BridgeRoot(f *ferry.Ferry) {
	app.Bridge(f, "")
}

// Bridge registers all of the application's controller's actions
// with a Ferry instance.
func (app *Application) Bridge(f *ferry.Ferry, namespace string) {
	// ensure namespace starts with a slash
	// and does not end on a slash
	http_bridge.NormalizeNamespace(namespace)

	for resource, controller := range app.controllers {
		prefix := namespace + "/" + resource.JSONAPIName()

		if len(controller.indexHandlers) > 0 {
			f.GET(prefix, controller.indexHandlers.toFerry(app, controller))
		}
		if len(controller.showHandlers) > 0 {
			f.GET(prefix+"/{id}", controller.showHandlers.toFerry(app, controller))
		}
		if len(controller.createHandlers) > 0 {
			f.POST(prefix, controller.createHandlers.toFerry(app, controller))
		}
		if len(controller.updateHandlers) > 0 {
			f.PATCH(prefix+"/{id}", controller.updateHandlers.toFerry(app, controller))
		}
		if len(controller.deleteHandlers) > 0 {
			f.DELETE(prefix+"/{id}", controller.deleteHandlers.toFerry(app, controller))
		}

		for route, handlers := range controller.customHandlers {
			if len(handlers) > 0 {
				f.Handle(route.method, prefix+route.path, handlers.toFerry(app, controller))
			}
		}
	}
}

// ToFerry creates a new Ferry instance hosting the Application
// under the given namespace.
func (app *Application) ToFerry(namespace string) *ferry.Ferry {
	f := ferry.New()
	app.Bridge(f, namespace)
	return f
}

// ServeHTTP serves the Application via HTTP under the given namespace.
// This is a blocking method.
func (app *Application) ServeHTTP(addr string, namespace string) error {
	mux := http.NewServeMux()
	http_bridge.BridgeRoot(app.ToFerry(namespace), mux)
	return http.ListenAndServe(addr, mux)
}
