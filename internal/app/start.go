package app

func Start(app *App) {
	app.cachePreheat.Warm()
	app.orderTimeout.Start()
	app.orderCreate.Start()
}
