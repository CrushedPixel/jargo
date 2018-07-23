package jargo

import (
	"context"
	"github.com/crushedpixel/ferry"
	"github.com/crushedpixel/http_bridge"
	"net/http"
)

// ensureAppRunning is a ferry handler function that panics if
// the app is not currently running.
func (app *Application) ensureAppRunning(r *ferry.Request) ferry.Response {
	if !app.running {
		panic(errAppNotRunning)
	}
	return nil
}

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

		// register custom routes first, to try and match them
		// before the catch-all (/{id}) routes
		for route, handlers := range controller.customHandlers {
			if len(handlers) > 0 {
				f.Handle(route.method, prefix+route.path, app.ensureAppRunning, handlers.toFerry(app, controller))
			}
		}

		// TODO: use ensureAppRunning as a router-level middleware once ferry#1 is resolved
		if len(controller.indexHandlers) > 0 {
			f.GET(prefix, app.ensureAppRunning, controller.indexHandlers.toFerry(app, controller))
		}
		if len(controller.showHandlers) > 0 {
			f.GET(prefix+"/{id}", app.ensureAppRunning, controller.showHandlers.toFerry(app, controller))
		}
		if len(controller.createHandlers) > 0 {
			f.POST(prefix, app.ensureAppRunning, controller.createHandlers.toFerry(app, controller))
		}
		if len(controller.updateHandlers) > 0 {
			f.PATCH(prefix+"/{id}", app.ensureAppRunning, controller.updateHandlers.toFerry(app, controller))
		}
		if len(controller.deleteHandlers) > 0 {
			f.DELETE(prefix+"/{id}", app.ensureAppRunning, controller.deleteHandlers.toFerry(app, controller))
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

// ServeHTTP serves the Application via HTTP under the given namespace until ctx is done.
// This is a blocking method.
func (app *Application) ServeHTTP(addr string, namespace string, ctx context.Context) error {
	mux := http.NewServeMux()
	http_bridge.BridgeRoot(app.ToFerry(namespace), mux)

	errorChan := make(chan error, 1)

	go func() {
		app.Run(ctx)
	}()

	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errorChan <- err
		}
	}()

	select {
	case err := <-errorChan:
		return err
	case <-ctx.Done():
		return nil
	}
}
